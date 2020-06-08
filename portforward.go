package portforward

import (
	"context"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// DialContext returns a connection to the specified cluster/namespace/pod/port.
func DialContext(ctx context.Context, logger logr.Logger, restconfig *rest.Config, namespace, pod, port string) (net.Conn, error) {

	spdyConn, err := dialSpdy(ctx, restconfig, "/api/v1/namespaces/"+namespace+"/pods/"+pod+"/portforward")
	if err != nil {
		return nil, errors.Wrap(err, "dialSpdy failed")
	}

	// Connect the error stream, r/o
	errorStream, err := spdyConn.CreateStream(http.Header{
		v1.StreamType:                 []string{v1.StreamTypeError},
		v1.PortHeader:                 []string{port},
		v1.PortForwardRequestIDHeader: []string{"0"},
	})
	if err != nil {
		spdyConn.Close()
		return nil, err
	}
	errorStream.Close() // this actually means CloseWrite()

	// Connect the data stream, r/w
	dataStream, err := spdyConn.CreateStream(http.Header{
		v1.StreamType:                 []string{v1.StreamTypeData},
		v1.PortHeader:                 []string{port},
		v1.PortForwardRequestIDHeader: []string{"0"},
	})
	if err != nil {
		spdyConn.Close()
		return nil, err
	}

	return newStreamConn(logger, spdyConn, dataStream, errorStream), nil
}
