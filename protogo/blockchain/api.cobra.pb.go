// Code generated by protoc-gen-cobra.
// source: proto/api.proto
// DO NOT EDIT!

/*
Package blockchain is a generated protocol buffer package.

It is generated from these files:
	proto/api.proto

It has these top-level commands:
	NodeClientCommand
*/

package blockchain

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	cobra "github.com/spf13/cobra"
	context "golang.org/x/net/context"
	credentials "google.golang.org/grpc/credentials"
	envconfig "github.com/kelseyhightower/envconfig"
	filepath "path/filepath"
	grpc "google.golang.org/grpc"
	io "io"
	iocodec "github.com/fiorix/protoc-gen-cobra/iocodec"
	ioutil "io/ioutil"
	json "encoding/json"
	log "log"
	net "net"
	oauth "google.golang.org/grpc/credentials/oauth"
	oauth2 "golang.org/x/oauth2"
	os "os"
	pflag "github.com/spf13/pflag"
	template "text/template"
	time "time"
	tls "crypto/tls"
	x509 "crypto/x509"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ cobra.Command
var _ context.Context
var _ credentials.AuthInfo
var _ envconfig.Decoder
var _ filepath.WalkFunc
var _ grpc.ClientConn
var _ io.Reader
var _ iocodec.Encoder
var _ = ioutil.Discard
var _ json.Encoder
var _ log.Logger
var _ net.IP
var _ oauth.TokenSource
var _ oauth2.Token
var _ os.File
var _ pflag.FlagSet
var _ template.Template
var _ time.Time
var _ tls.Config
var _ x509.Certificate

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

var _DefaultNodeClientCommandConfig = _NewNodeClientCommandConfig()

type _NodeClientCommandConfig struct {
	ServerAddr         string        `envconfig:"SERVER_ADDR" default:"localhost:8080"`
	RequestFile        string        `envconfig:"REQUEST_FILE"`
	PrintSampleRequest bool          `envconfig:"PRINT_SAMPLE_REQUEST"`
	ResponseFormat     string        `envconfig:"RESPONSE_FORMAT" default:"json"`
	Timeout            time.Duration `envconfig:"TIMEOUT" default:"10s"`
	TLS                bool          `envconfig:"TLS"`
	ServerName         string        `envconfig:"TLS_SERVER_NAME"`
	InsecureSkipVerify bool          `envconfig:"TLS_INSECURE_SKIP_VERIFY"`
	CACertFile         string        `envconfig:"TLS_CA_CERT_FILE"`
	CertFile           string        `envconfig:"TLS_CERT_FILE"`
	KeyFile            string        `envconfig:"TLS_KEY_FILE"`
	AuthToken          string        `envconfig:"AUTH_TOKEN"`
	AuthTokenType      string        `envconfig:"AUTH_TOKEN_TYPE" default:"Bearer"`
	JWTKey             string        `envconfig:"JWT_KEY"`
	JWTKeyFile         string        `envconfig:"JWT_KEY_FILE"`
}

func _NewNodeClientCommandConfig() *_NodeClientCommandConfig {
	c := &_NodeClientCommandConfig{}
	envconfig.Process("", c)
	return c
}

func (o *_NodeClientCommandConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ServerAddr, "server-addr", "s", o.ServerAddr, "server address in form of host:port")
	fs.StringVarP(&o.RequestFile, "request-file", "f", o.RequestFile, "client request file (must be json, yaml, or xml); use \"-\" for stdin + json")
	fs.BoolVarP(&o.PrintSampleRequest, "print-sample-request", "p", o.PrintSampleRequest, "print sample request file and exit")
	fs.StringVarP(&o.ResponseFormat, "response-format", "o", o.ResponseFormat, "response format (json, prettyjson, yaml, or xml)")
	fs.DurationVar(&o.Timeout, "timeout", o.Timeout, "client connection timeout")
	fs.BoolVar(&o.TLS, "tls", o.TLS, "enable tls")
	fs.StringVar(&o.ServerName, "tls-server-name", o.ServerName, "tls server name override")
	fs.BoolVar(&o.InsecureSkipVerify, "tls-insecure-skip-verify", o.InsecureSkipVerify, "INSECURE: skip tls checks")
	fs.StringVar(&o.CACertFile, "tls-ca-cert-file", o.CACertFile, "ca certificate file")
	fs.StringVar(&o.CertFile, "tls-cert-file", o.CertFile, "client certificate file")
	fs.StringVar(&o.KeyFile, "tls-key-file", o.KeyFile, "client key file")
	fs.StringVar(&o.AuthToken, "auth-token", o.AuthToken, "authorization token")
	fs.StringVar(&o.AuthTokenType, "auth-token-type", o.AuthTokenType, "authorization token type")
	fs.StringVar(&o.JWTKey, "jwt-key", o.JWTKey, "jwt key")
	fs.StringVar(&o.JWTKeyFile, "jwt-key-file", o.JWTKeyFile, "jwt key file")
}

var NodeClientCommand = &cobra.Command{
	Use: "node",
}

func _DialNode() (*grpc.ClientConn, NodeClient, error) {
	cfg := _DefaultNodeClientCommandConfig
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(cfg.Timeout),
	}
	if cfg.TLS {
		tlsConfig := &tls.Config{}
		if cfg.InsecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		if cfg.CACertFile != "" {
			cacert, err := ioutil.ReadFile(cfg.CACertFile)
			if err != nil {
				return nil, nil, fmt.Errorf("ca cert: %v", err)
			}
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(cacert)
			tlsConfig.RootCAs = certpool
		}
		if cfg.CertFile != "" {
			if cfg.KeyFile == "" {
				return nil, nil, fmt.Errorf("missing key file")
			}
			pair, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
			if err != nil {
				return nil, nil, fmt.Errorf("cert/key: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{pair}
		}
		if cfg.ServerName != "" {
			tlsConfig.ServerName = cfg.ServerName
		} else {
			addr, _, _ := net.SplitHostPort(cfg.ServerAddr)
			tlsConfig.ServerName = addr
		}
		//tlsConfig.BuildNameToCertificate()
		cred := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	if cfg.AuthToken != "" {
		cred := oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: cfg.AuthToken,
			TokenType:   cfg.AuthTokenType,
		})
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if cfg.JWTKey != "" {
		cred, err := oauth.NewJWTAccessFromKey([]byte(cfg.JWTKey))
		if err != nil {
			return nil, nil, fmt.Errorf("jwt key: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	if cfg.JWTKeyFile != "" {
		cred, err := oauth.NewJWTAccessFromFile(cfg.JWTKeyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("jwt key file: %v", err)
		}
		opts = append(opts, grpc.WithPerRPCCredentials(cred))
	}
	conn, err := grpc.Dial(cfg.ServerAddr, opts...)
	if err != nil {
		return nil, nil, err
	}
	return conn, NewNodeClient(conn), nil
}

type _NodeRoundTripFunc func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error

func _NodeRoundTrip(sample interface{}, fn _NodeRoundTripFunc) error {
	cfg := _DefaultNodeClientCommandConfig
	var em iocodec.EncoderMaker
	var ok bool
	if cfg.ResponseFormat == "" {
		em = iocodec.DefaultEncoders["json"]
	} else {
		em, ok = iocodec.DefaultEncoders[cfg.ResponseFormat]
		if !ok {
			return fmt.Errorf("invalid response format: %q", cfg.ResponseFormat)
		}
	}
	if cfg.PrintSampleRequest {
		return em.NewEncoder(os.Stdout).Encode(sample)
	}
	var d iocodec.Decoder
	if cfg.RequestFile == "" || cfg.RequestFile == "-" {
		d = iocodec.DefaultDecoders["json"].NewDecoder(os.Stdin)
	} else {
		f, err := os.Open(cfg.RequestFile)
		if err != nil {
			return fmt.Errorf("request file: %v", err)
		}
		defer f.Close()
		ext := filepath.Ext(cfg.RequestFile)
		if len(ext) > 0 && ext[0] == '.' {
			ext = ext[1:]
		}
		dm, ok := iocodec.DefaultDecoders[ext]
		if !ok {
			return fmt.Errorf("invalid request file format: %q", ext)
		}
		d = dm.NewDecoder(f)
	}
	conn, client, err := _DialNode()
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(client, d, em.NewEncoder(os.Stdout))
}

var _NodeDiscoverClientCommand = &cobra.Command{
	Use:  "discover",
	Long: "Discover client\n\nYou can use environment variables with the same name of the command flags.\nAll caps and s/-/_, e.g. SERVER_ADDR.",
	Example: `
Save a sample request to a file (or refer to your protobuf descriptor to create one):
	discover -p > req.json

Submit request using file:
	discover -f req.json

Authenticate using the Authorization header (requires transport security):
	export AUTH_TOKEN=your_access_token
	export SERVER_ADDR=api.example.com:443
	echo '{json}' | discover --tls`,
	Run: func(cmd *cobra.Command, args []string) {
		var v DiscoverRequest
		err := _NodeRoundTrip(v, func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error {

			err := in.Decode(&v)
			if err != nil {
				return err
			}

			resp, err := cli.Discover(context.Background(), &v)

			if err != nil {
				return err
			}

			return out.Encode(resp)

		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	NodeClientCommand.AddCommand(_NodeDiscoverClientCommand)
	_DefaultNodeClientCommandConfig.AddFlags(_NodeDiscoverClientCommand.Flags())
}

var _NodeGetStateClientCommand = &cobra.Command{
	Use:  "getstate",
	Long: "GetState client\n\nYou can use environment variables with the same name of the command flags.\nAll caps and s/-/_, e.g. SERVER_ADDR.",
	Example: `
Save a sample request to a file (or refer to your protobuf descriptor to create one):
	getstate -p > req.json

Submit request using file:
	getstate -f req.json

Authenticate using the Authorization header (requires transport security):
	export AUTH_TOKEN=your_access_token
	export SERVER_ADDR=api.example.com:443
	echo '{json}' | getstate --tls`,
	Run: func(cmd *cobra.Command, args []string) {
		var v GetStateRequest
		err := _NodeRoundTrip(v, func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error {

			err := in.Decode(&v)
			if err != nil {
				return err
			}

			resp, err := cli.GetState(context.Background(), &v)

			if err != nil {
				return err
			}

			return out.Encode(resp)

		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	NodeClientCommand.AddCommand(_NodeGetStateClientCommand)
	_DefaultNodeClientCommandConfig.AddFlags(_NodeGetStateClientCommand.Flags())
}

var _NodeShareChainClientCommand = &cobra.Command{
	Use:  "sharechain",
	Long: "ShareChain client\n\nYou can use environment variables with the same name of the command flags.\nAll caps and s/-/_, e.g. SERVER_ADDR.",
	Example: `
Save a sample request to a file (or refer to your protobuf descriptor to create one):
	sharechain -p > req.json

Submit request using file:
	sharechain -f req.json

Authenticate using the Authorization header (requires transport security):
	export AUTH_TOKEN=your_access_token
	export SERVER_ADDR=api.example.com:443
	echo '{json}' | sharechain --tls`,
	Run: func(cmd *cobra.Command, args []string) {
		var v ShareChainRequest
		err := _NodeRoundTrip(v, func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error {

			err := in.Decode(&v)
			if err != nil {
				return err
			}

			resp, err := cli.ShareChain(context.Background(), &v)

			if err != nil {
				return err
			}

			return out.Encode(resp)

		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	NodeClientCommand.AddCommand(_NodeShareChainClientCommand)
	_DefaultNodeClientCommandConfig.AddFlags(_NodeShareChainClientCommand.Flags())
}

var _NodeShareTxClientCommand = &cobra.Command{
	Use:  "sharetx",
	Long: "ShareTx client\n\nYou can use environment variables with the same name of the command flags.\nAll caps and s/-/_, e.g. SERVER_ADDR.",
	Example: `
Save a sample request to a file (or refer to your protobuf descriptor to create one):
	sharetx -p > req.json

Submit request using file:
	sharetx -f req.json

Authenticate using the Authorization header (requires transport security):
	export AUTH_TOKEN=your_access_token
	export SERVER_ADDR=api.example.com:443
	echo '{json}' | sharetx --tls`,
	Run: func(cmd *cobra.Command, args []string) {
		var v ShareTxRequest
		err := _NodeRoundTrip(v, func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error {

			err := in.Decode(&v)
			if err != nil {
				return err
			}

			resp, err := cli.ShareTx(context.Background(), &v)

			if err != nil {
				return err
			}

			return out.Encode(resp)

		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	NodeClientCommand.AddCommand(_NodeShareTxClientCommand)
	_DefaultNodeClientCommandConfig.AddFlags(_NodeShareTxClientCommand.Flags())
}

var _NodeGetCreditClientCommand = &cobra.Command{
	Use:  "getcredit",
	Long: "GetCredit client\n\nYou can use environment variables with the same name of the command flags.\nAll caps and s/-/_, e.g. SERVER_ADDR.",
	Example: `
Save a sample request to a file (or refer to your protobuf descriptor to create one):
	getcredit -p > req.json

Submit request using file:
	getcredit -f req.json

Authenticate using the Authorization header (requires transport security):
	export AUTH_TOKEN=your_access_token
	export SERVER_ADDR=api.example.com:443
	echo '{json}' | getcredit --tls`,
	Run: func(cmd *cobra.Command, args []string) {
		var v GetCreditRequest
		err := _NodeRoundTrip(v, func(cli NodeClient, in iocodec.Decoder, out iocodec.Encoder) error {

			err := in.Decode(&v)
			if err != nil {
				return err
			}

			resp, err := cli.GetCredit(context.Background(), &v)

			if err != nil {
				return err
			}

			return out.Encode(resp)

		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	NodeClientCommand.AddCommand(_NodeGetCreditClientCommand)
	_DefaultNodeClientCommandConfig.AddFlags(_NodeGetCreditClientCommand.Flags())
}
