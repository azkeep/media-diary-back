package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImportCSV(t *testing.T) {
	tests := []struct {
		name               string
		filename           string
		content            string
		mockErr            error
		expectedStatusCode int
	}{
		{
			name:               "should return 200 OK when uploading valid CSV file",
			filename:           "data.csv",
			content:            "Date;Title,IsFinished\n01.01.2026;Matrix;true",
			mockErr:            nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should return 400 Bad Request when file extension is not csv",
			filename:           "data.txt",
			content:            "some text content",
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 400 Bad Request when file form field is missing",
			filename:           "",
			content:            "",
			mockErr:            nil,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should return 400 Bad Request when service import fails",
			filename:           "data.csv",
			content:            "invalid csv content",
			mockErr:            fmt.Errorf("csv parse error"),
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := setupTestServer(&mockService{err: tt.mockErr})

			req, err := createMultipartCSVRequest(tt.filename, tt.content)
			if err != nil {
				t.Fatalf("failed creating multipart request: %v", err)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			assertSingleJSONResponse[map[string]string](t, rec, tt.expectedStatusCode)
		})
	}
}
