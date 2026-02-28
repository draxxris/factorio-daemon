package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/instances/:name", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"name": c.Param("name")})
		})
		apiGroup.POST("/instances/:name/start", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "instance started"})
		})
	}
	return r
}

func TestRouting_InstanceGet(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/instances/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRouting_InstanceStart(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("POST", "/api/instances/test/start", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRouting_NotFound(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
