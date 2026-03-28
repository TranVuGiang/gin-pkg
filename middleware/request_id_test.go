package middleware_test

import (
	"net/http"
	"testing"

	"github.com/TranVuGiang/gin-pkg/middleware"
	"github.com/TranVuGiang/gin-pkg/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type RequestIDTestCase struct {
	name           string
	requestID      string
	expectedStatus int
	skipper        middleware.Skipper
}

func generateRequestIDTestCases() []RequestIDTestCase {
	return []RequestIDTestCase{
		{
			name:           "Valid Request ID",
			requestID:      uuid.New().String(),
			expectedStatus: http.StatusOK,
			skipper:        middleware.DefaultSkipper,
		},
		{
			name:           "Missing Request ID",
			requestID:      "",
			expectedStatus: http.StatusBadRequest,
			skipper:        middleware.DefaultSkipper,
		},
		{
			name:           "Invalid Request ID",
			requestID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			skipper:        middleware.DefaultSkipper,
		},
		{
			name:           "Skipped Middleware",
			requestID:      "",
			expectedStatus: http.StatusOK,
			skipper:        func(_ *gin.Context) bool { return true },
		},
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	testCases := generateRequestIDTestCases()
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c, w, req := testutil.SetupGinContext(t, &testutil.Options{
				Method: http.MethodPost,
				Path:   "/test",
			})

			if test.requestID != "" {
				req.Header.Set(middleware.HeaderXRequestID, test.requestID)
			}

			mw := middleware.RequestID(test.skipper)
			mw(c)

			if test.expectedStatus == http.StatusOK {
				require.False(t, c.IsAborted())

				if !test.skipper(c) {
					rid, exists := c.Get(middleware.ContextKeyRequestID)
					require.True(t, exists)
					require.Equal(t, test.requestID, rid)
				}
			} else {
				require.True(t, c.IsAborted())
				require.Equal(t, test.expectedStatus, w.Code)
			}
		})
	}
}
