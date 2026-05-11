package server

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/swaggest/swgui/v5emb"
)

func mountSwagger(r *gin.Engine, specPath string) error {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}

	r.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", spec)
	})

	ui := v5emb.New("API", "/openapi.json", "/docs/")
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/")
	})
	r.GET("/docs/*any", gin.WrapH(ui))
	return nil
}
