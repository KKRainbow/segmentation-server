package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"runtime"
	"io/ioutil"
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
	"os"
	"github.com/KKRainbow/segmentation-server/segmentation"
)

func getcurdir() (string, error) {
	//get absolute path, which will be used to locate html and js files
	_, filename, _, _ := runtime.Caller(1)
	dir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dir+"frontend/dist"); err != nil {
		t, err := os.Getwd()
		return t + "/", err
	}
	return dir + "/", nil
}

func main() {
	modelCache := make(map[string]*segmentation.Segmentation)
	dagCache := make(map[string]*segmentation.DAGBuilder)

	curDir, _ := getcurdir()
	fmt.Println("curdir: ", curDir)

	g := gin.Default()

	g.StaticFS("/", http.Dir(curDir+"frontend/dist"))

	g.OPTIONS("/segmentation", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})
	g.POST("/segmentation", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		payload, _ := ioutil.ReadAll(c.Request.Body)
		req := make(map[string]string)
		json.Unmarshal(payload, &req)

		fmt.Println("Get request:", req)

		model := req["model-file"]
		dict := req["dict-file"]
		phrase := req["phrase-file"]
		maxStep, _ := strconv.Atoi(req["max-step"])
		maxLength, _ := strconv.Atoi(req["max-length"])
		batch_size, _ := strconv.Atoi(req["batch-size"])
		lines := strings.Split(req["strings"], "\n")

		var ok bool
		var dagBuilder *segmentation.DAGBuilder
		dagCacheKey := fmt.Sprint(dict, phrase, maxStep)
		if dagBuilder, ok = dagCache[dagCacheKey]; !ok {
			dagBuilder = segmentation.NewDAGBuilder(dict, phrase, maxStep)
			dagCache[dagCacheKey] = dagBuilder
		}

		var seg *segmentation.Segmentation
		segCacheKey := fmt.Sprint(model, )
		if seg, ok = modelCache[segCacheKey]; !ok {
			var err error
			seg, err = segmentation.NewSegmentation(model, dict, maxLength, batch_size, dagBuilder)
			modelCache[segCacheKey] = seg
			if err != nil {
				c.Error(err)
				return
			}
		}

		res, err := seg.SegmentLine(lines)
		if err != nil {
			c.Error(err)
		}

		c.JSON(http.StatusOK, res)
	})

	g.Run(":8888")
}

