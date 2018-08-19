// Code generated by protoc-gen-twirp_metrics v0.0.1, DO NOT EDIT.
// source: test.proto

package proto

import "context"
import "github.com/Xe/ln"

// HelloWorldLogging is a middleware for HelloWorld that logs all usage of the methods
type HelloWorldLogging struct {
	next HelloWorld
}

func NewHelloWorldLogging(next HelloWorld) HelloWorld {
	var result HelloWorldLogging
	result.next = next
	return result
}

func (i HelloWorldLogging) Speak(ctx context.Context, input *Words) (result *Words, err error) {
	ctx = ln.WithF(ctx, ln.F{
		"twirp_package": "us.xeserv.api",
		"twirp_service": "HelloWorld",
		"twirp_method": "Speak",
	})
	result, err = i.next.Speak(ctx, input)
	if err != nil {
		ln.Error(ctx, err, input)
	}
	return
}

// F ields for logging.
func (i Words) F() ln.F {
	return ln.F{
		"words_message": i.Message,
	}
}


