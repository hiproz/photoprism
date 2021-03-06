package api

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// GET /api/v1/thumbnails/:hash/:type
//
// Parameters:
//   hash: string The file hash as returned by the search API
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func GetThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/thumbnails/:hash/:type", func(c *gin.Context) {
		fileHash := c.Param("hash")
		typeName := c.Param("type")

		thumbType, ok := photoprism.ThumbnailTypes[typeName]

		if !ok {
			log.Errorf("invalid type: %s", typeName)
			c.Data(400, "image/svg+xml", photoIconSvg)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		file, err := search.FindFileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(404, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		if thumbnail, err := photoprism.ThumbnailFromFile(fileName, file.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				downloadFileName := file.DownloadFileName()

				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(400, "image/svg+xml", photoIconSvg)
		}
	})
}
