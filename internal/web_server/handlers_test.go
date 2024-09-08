package webserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T){
	//подготовка
	stor := storage.New()
	
	name := "name"
	value := 1.14
	itemType1 := storage.TYPE1
	itemType2 := storage.TYPE2

	item1 := storage.Metric{
		Type: itemType1,
		Name: name,
		Value: value,
	}

	item2 := storage.Metric{
		Type: itemType2,
		Name: name,
		Value: value,
	}

	rSuccess1 := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx1 := context.WithValue(rSuccess1.Context(), STORAGE_KEY, stor)
	ctx1 = context.WithValue(ctx1, METRIC_KEY, &item1)
	rSuccess1 = rSuccess1.WithContext(ctx1)

	rSuccess2 := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx2 := context.WithValue(rSuccess2.Context(), STORAGE_KEY, stor)
	ctx2 = context.WithValue(ctx2, METRIC_KEY, &item2)
	rSuccess2 = rSuccess2.WithContext(ctx2)

	rWrongMethod := httptest.NewRequest(http.MethodGet, "/", nil)
	ctxwm := context.WithValue(rWrongMethod.Context(), STORAGE_KEY, stor)
	ctxwm = context.WithValue(ctxwm,  METRIC_KEY, &item2)
	rWrongMethod = rWrongMethod.WithContext(ctxwm)

	rWithoutStor := httptest.NewRequest(http.MethodPost, "/", nil)
	ctxws := context.WithValue(rWithoutStor.Context(), METRIC_KEY, &item1)
	rWithoutStor = rWithoutStor.WithContext(ctxws)

	tests := []struct{
		req *http.Request
		waitErr bool
		message string
	}{
		{rSuccess1, false, "type 1 success"},
		{rSuccess2, false, "type 2 success"},
		{rWrongMethod, true, "wrong method"},
		{rWithoutStor, true, "without store in context"},
	}

	

	for _, tt := range tests{
		w := httptest.NewRecorder()

		Update(w, tt.req)

		if tt.waitErr{
			assert.NotEqual(t, http.StatusOK, w.Result().StatusCode, tt.message)
		} else {
			assert.Equal(t, http.StatusOK, w.Result().StatusCode, tt.message)
		}
	}
}