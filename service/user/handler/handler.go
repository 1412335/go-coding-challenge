package handler

import (
	"context"
	"crypto/tls"
	"mime"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rakyll/statik/fs"
	"github.com/unrolled/secure"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/errors"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/utils"

	// Static files
	_ "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user/statik"
	// requirement for using error details and want to marshal them correctly to JSON.
	// https://jbrandhorst.com/post/grpc-errors/
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Handler struct {
	logger log.Factory
	config *configs.ServiceConfig
}

func NewHandler(config *configs.ServiceConfig) *Handler {
	return &Handler{
		logger: log.With(zap.String("gateway", "gin")),
		config: config,
	}
}

// isPermanentHTTPHeader checks whether hdr belongs to the list of
// permanent request headers maintained by IANA.
// http://www.iana.org/assignments/message-headers/message-headers.xml
// From https://github.com/grpc-ecosystem/grpc-gateway/blob/7a2a43655ccd9a488d423ea41a3fc723af103eda/runtime/context.go#L157
func (h *Handler) isPermanentHTTPHeader(hdr string) bool {
	switch hdr {
	case
		"Accept",
		"Accept-Charset",
		"Accept-Language",
		"Accept-Ranges",
		"Authorization",
		"Cache-Control",
		"Content-Type",
		"Cookie",
		"Date",
		"Expect",
		"From",
		"Host",
		"If-Match",
		"If-Modified-Since",
		"If-None-Match",
		"If-Schedule-Tag-Match",
		"If-Unmodified-Since",
		"Max-Forwards",
		"Origin",
		"Pragma",
		"Referer",
		"User-Agent",
		"Via",
		"Warning":
		return true
	}
	return false
}

// isReserved returns whether the key is reserved by gRPC.
func (h *Handler) isReserved(key string) bool {
	return strings.HasPrefix(key, "Grpc-")
}

// incomingHeaderMatcher converts an HTTP header name on http.Request to
// grpc metadata. Permanent headers (i.e. User-Agent) are prepended with
// "grpc-gateway". Headers that start with start with "Grpc-" (reserved
// by grpc) are prepended with "X-". Other headers are forwarded as is.
func (h *Handler) incomingHeaderMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if h.isPermanentHTTPHeader(key) {
		return runtime.MetadataPrefix + key, true
	}
	if h.isReserved(key) {
		return "X-" + key, true
	}

	// The Istio service mesh dislikes when you pass the Content-Length header
	if key == "Content-Length" {
		return "", false
	}

	return key, true
}

// outgoingHeaderMatcher transforms outgoing metadata into HTTP headers.
// We return any response metadata as is.
func (h *Handler) outgoingHeaderMatcher(metadata string) (string, bool) {
	return metadata, true
}

// init gin router
func (h *Handler) initRouter(handler http.Handler) *gin.Engine {
	if os.Getenv("GOENV") != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	secureMiddleware := secure.New(secure.Options{
		AllowedHosts: []string{
			"*",
		},
		AllowedHostsAreRegex:  true,
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLRedirect:           false,
		SSLHost:               "",
		SSLProxyHeaders:       map[string]string{},
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self' data:; font-src 'self' data:",
		PublicKey:             `pin-sha256="base64+primary=="; pin-sha256="base64+backup=="; max-age=5184000; includeSubdomains; report-uri="http://localhost:8082"`,
	})

	secureFunc := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := secureMiddleware.Process(c.Writer, c.Request)

			// If there was an error, do not continue.
			if err != nil {
				h.logger.For(c).Error("Process", zap.Error(err))
				c.Abort()
				return
			}

			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				h.logger.For(c).Error("Write status", zap.Int("status", status), zap.Error(err))
				c.Abort()
			}
		}
	}()

	r := gin.Default()
	r.Use(secureFunc)

	if err := serveOpenAPI(r); err != nil {
		h.logger.Bg().Error("Serve OpenAPI", zap.Error(err))
	}

	api := r.Group("/api/v1")
	api.Any("/*any", gin.WrapH(handler))

	return r
}

// serveOpenAPI serves an OpenAPI UI on /openapi-ui/
// Adapted from https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L63
func serveOpenAPI(r *gin.Engine) error {
	if err := mime.AddExtensionType(".svg", "image/svg+xml"); err != nil {
		return err
	}
	statikFS, err := fs.New()
	if err != nil {
		return err
	}
	// Expose files in static on <host>/openapi-ui
	// fileServer := http.FileServer(statikFS)
	prefix := "/openapi-ui/"
	r.StaticFS(prefix, statikFS)
	return nil
}

func (h *Handler) loadClientTLSCredentials() (credentials.TransportCredentials, error) {
	config, err := utils.LoadClientTLSConfig(h.config.TLSCert.CACert)
	if err != nil {
		return nil, err
	}
	// config.ServerName = h.addr
	config.InsecureSkipVerify = true
	// Create the credentials and return it
	return credentials.NewTLS(config), nil
}

func (h *Handler) loadServerTLSCredentials() (*tls.Config, error) {
	config, err := utils.LoadServerTLSConfig(h.config.TLSCert.CertPem, h.config.TLSCert.KeyPem)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// run grpc-gateway
func (h *Handler) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// custom http error
	runtime.HTTPError = errors.CustomHTTPError

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(h.incomingHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(h.outgoingHeaderMatcher),
		// runtime.WithMarshalerOption(runtime.MIMEWildcard, &gateway.JSONPb{
		// 	OrigName:     true,
		// 	EmitDefaults: false,
		// 	Indent:       "  ",
		// }),
		// // This is necessary to get error details properly
		// // marshaled in unary requests.
		// runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)

	gRPCHost := net.JoinHostPort("0.0.0.0", strconv.Itoa(h.config.GRPC.Port))

	// gRPC client options
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// insecure
	if h.config.EnableTLS && h.config.TLSCert != nil {
		creds, err := h.loadClientTLSCredentials()
		if err != nil {
			h.logger.For(ctx).Error("Load client TLS credentials", zap.Error(err))
			return err
		}
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			// grpc.WithBlock(),
		}
	}

	callOptions := []grpc.CallOption{}
	if h.config.GRPC.MaxCallRecvMsgSize > 0 {
		callOptions = append(callOptions, grpc.MaxCallRecvMsgSize(h.config.GRPC.MaxCallRecvMsgSize))
	}
	if h.config.GRPC.MaxCallSendMsgSize > 0 {
		callOptions = append(callOptions, grpc.MaxCallSendMsgSize(h.config.GRPC.MaxCallSendMsgSize))
	}
	if len(callOptions) > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(callOptions...))
	}

	// register handler
	err := pb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		mux,
		gRPCHost,
		opts,
	)
	if err != nil {
		h.logger.For(ctx).Error("Register gateway", zap.Error(err))
		return err
	}

	// proxy address
	addr := ":" + strconv.Itoa(h.config.Proxy.Port)
	// router
	router := h.initRouter(mux)
	// http server
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// insecure
	if h.config.EnableTLS && h.config.TLSCert != nil {
		tlsConfig, err := h.loadServerTLSCredentials()
		if err != nil {
			h.logger.For(ctx).Error("Load http server TLS credentials", zap.Error(err))
			return err
		}
		srv.TLSConfig = tlsConfig
	}

	// graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signals
		h.logger.For(ctx).Info("Received signal", zap.Any("signal", sig))
		shutdown, can := context.WithTimeout(ctx, 10*time.Second)
		srv.Shutdown(shutdown)
		defer can()
	}()

	h.logger.For(ctx).Info("Serving gRPC-Gateway on", zap.String("addr", "http://"+addr), zap.String("openapi", "http://"+addr+"/openapi-ui/"))
	// run gateway
	if h.config.EnableTLS && h.config.TLSCert != nil {
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}
