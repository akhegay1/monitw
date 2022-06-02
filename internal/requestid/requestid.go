package requestid

import (
	"net/http"

	"golang.org/x/net/context"

	log "monitw/internal/logger"
	"monitw/pkg/mwutils"

	"github.com/google/uuid"
)

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
/*type ContextKey string // can be unexported

// ContextKeyRequestID is the ContextKey for RequestID
const ContextKeyRequestID ContextKey = "requestID" // can be unexported
*/

// AttachRequestID will attach a brand new request ID to a http request
func AssignRequestID(ctx context.Context) context.Context {

	reqID := uuid.New()
	log.Println(mwutils.FuncName(), "reqID UUid="+reqID.String())
	return context.WithValue(ctx, "requestID", reqID.String())
}

// GetRequestID will get reqID from a http request and return it as a string
func GetRequestID(ctx context.Context) string {

	reqID := ctx.Value("requestID")
	//rIdStr, ok := ctx.Value(ContextKeyRequestID).(string)
	//fmt.Println("ok=", ok)
	//log.Logger.Println("rIdStr=" + rIdStr)

	if ret, ok := reqID.(string); ok {
		return ret
	}

	return ""
}

//#region middlewares

func ReqIDMiddleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx2 := AssignRequestID(ctx)

		r = r.WithContext(ctx2)
		//rIdStr, ok := ctx2.Value(ContextKeyRequestID).(string)
		//log.Logger.Println("ok=", ok)
		//log.Logger.Println("rIdStr=" + rIdStr)

		log.PrintlnArgs(mwutils.FuncName(), "Incomming request ", r.Method, r.RequestURI, r.RemoteAddr)

		next.ServeHTTP(w, r)

		log.Println(mwutils.FuncName(), "Finished handling http req")
	})
}
