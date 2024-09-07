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
	value := "value"
	itemType1 := storage.TYPE1
	itemType2 := storage.TYPE2

	rSuccess1 := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx1 := context.WithValue(rSuccess1.Context(), STORAGE_KEY, stor)
	ctx1 = context.WithValue(ctx1, TYPE_KEY, itemType1)
	ctx1 = context.WithValue(ctx1, NAME_KEY, name)
	ctx1 = context.WithValue(ctx1, VALUE_KEY, value)
	rSuccess1 = rSuccess1.WithContext(ctx1)

	rSuccess2 := httptest.NewRequest(http.MethodPost, "/", nil)
	ctx2 := context.WithValue(rSuccess2.Context(), STORAGE_KEY, stor)
	ctx2 = context.WithValue(ctx2, TYPE_KEY, itemType2)
	ctx2 = context.WithValue(ctx2, NAME_KEY, name)
	ctx2 = context.WithValue(ctx2, VALUE_KEY, value)
	rSuccess2 = rSuccess2.WithContext(ctx2)

	rWrongMethod := httptest.NewRequest(http.MethodGet, "/", nil)
	ctxwm := context.WithValue(rWrongMethod.Context(), STORAGE_KEY, stor)
	ctxwm = context.WithValue(ctxwm, TYPE_KEY, itemType2)
	ctxwm = context.WithValue(ctxwm, NAME_KEY, name)
	ctxwm = context.WithValue(ctxwm, VALUE_KEY, value)
	rWrongMethod = rWrongMethod.WithContext(ctxwm)

	rWithoutStor := httptest.NewRequest(http.MethodPost, "/", nil)
	ctxws := context.WithValue(rWithoutStor.Context(), TYPE_KEY, itemType2)
	ctxws = context.WithValue(ctxws, NAME_KEY, name)
	ctxws = context.WithValue(ctxws, VALUE_KEY, value)
	rWithoutStor = rWithoutStor.WithContext(ctxws)

	rWrongType := httptest.NewRequest(http.MethodPost, "/", nil)
	ctxwt := context.WithValue(rWrongType.Context(), STORAGE_KEY, stor)
	ctxwt = context.WithValue(ctxwt, TYPE_KEY, rSuccess1.PathValue("wrong"))
	ctxwt = context.WithValue(ctxwt, NAME_KEY, rSuccess1.PathValue(name))
	ctxwt = context.WithValue(ctxwt, VALUE_KEY, rSuccess1.PathValue(value))
	rWrongType = rWrongType.WithContext(ctxwt)


	tests := []struct{
		req *http.Request
		waitErr bool
		message string
	}{
		{rSuccess1, false, "type 1 success"},
		{rSuccess2, false, "type 2 success"},
		{rWrongMethod, true, "wrong method"},
		{rWithoutStor, true, "without store in context"},
		{rWrongType, true, "wrong metric type"},
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