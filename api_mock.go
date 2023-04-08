package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ApiMockData = Server{}

func main() {
	args := os.Args
	log.Println(args)

	var mockDataFile = ""
	if len(args) == 2 {
		mockDataFile = args[1]
	}
	initMockData(mockDataFile)

	var logFile = ""
	if len(args) == 3 {
		logFile = args[2]
	}
	setupLog(logFile)

	server := setupServer(ApiMockData)

	var err error
	if len(ApiMockData.KeyFile) > 0 && len(ApiMockData.CertFile) > 0 {
		err = server.ListenAndServeTLS(ApiMockData.CertFile, ApiMockData.KeyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		log.Fatal(err)
	}
}

func setupLog(l string) {
	if len(l) == 0 {
		now := fmt.Sprintf("%s", time.Now().UTC())
		now = strings.Replace(now, " ", "_", -1)
		now = strings.Replace(now, ":", "-", -1)
		l = "gin-" + ApiMockData.Name + "-" + now + ".log"

		lfs, err := filepath.Glob("gin-" + ApiMockData.Name + "-*")
		if err != nil {
			log.Fatalf("glob the log file [%s] error: %v\n", "gin-"+ApiMockData.Name, err)
		}
		if len(lfs) > 10 {
			sort.Sort(sort.Reverse(sort.StringSlice(lfs)))
			for _, value := range lfs[10:] {
				_ = os.Remove(value)
			}
		}
	}

	f, _ := os.Create(l)
	gin.DefaultWriter = io.MultiWriter(f)
}

func setupServer(server Server) *http.Server {
	if server.ReadTimeout <= 0 {
		server.ReadTimeout = 5
	}
	if server.WriteTimeout <= 0 {
		server.WriteTimeout = 10
	}
	return &http.Server{
		Addr:         server.Address,
		Handler:      setupHandler(server),
		ReadTimeout:  time.Duration(server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(server.WriteTimeout) * time.Second,
	}
}

var apiMap = make(map[string]Api)

func setupHandler(server Server) http.Handler {
	e := gin.New()
	e.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	e.Use(gin.Recovery())

	for _, api := range server.Apis {
		if len(api.Response.ContentType) == 0 {
			api.Response.ContentType = "application/json"
		}
		apiMap[api.Request.Method+":"+api.Request.Url] = api

		e.Handle(api.Request.Method, api.Request.Url, func(c *gin.Context) {
			a, varsMap := setupTemplateVars(c)

			for _, h := range a.Response.Headers {
				c.Header(h.Key, decoder(h.Value, varsMap))
			}

			c.Data(a.Response.Status, a.Response.ContentType, []byte(decoder(a.Response.Body, varsMap)))
		})
	}

	return e
}

func setupTemplateVars(c *gin.Context) (Api, map[string]interface{}) {
	apiKey := c.Request.Method + ":" + c.FullPath()
	a := apiMap[apiKey]

	var varsMap = make(map[string]interface{})
	pathsMap := make(map[string]interface{})
	for _, v := range a.Request.Metadata.PathVars {
		pathsMap[v] = c.Param(v)
	}
	varsMap["path_vars"] = pathsMap

	headersMap := make(map[string]interface{})
	for _, k := range a.Request.Metadata.HeaderKeys {
		headersMap[k] = c.GetHeader(k)
	}
	varsMap["header_keys"] = headersMap

	paramsMap := make(map[string]interface{})
	for _, param := range a.Request.Metadata.QueryParams {
		paramsMap[param] = c.Query(param)
	}
	varsMap["query_params"] = paramsMap

	formsMap := make(map[string]interface{})
	for _, v := range a.Request.Metadata.FormVars {
		formsMap[v] = c.PostForm(v)
	}
	varsMap["form_vars"] = formsMap

	if len(a.Request.Metadata.JsonBodyKeys) > 0 {
		body, err := c.GetRawData()
		if err != nil {
			log.Printf("read body error, %s\n", apiKey)
		} else {
			if len(body) > 0 {
				var b interface{}
				err = json.Unmarshal(body, &b)
				if err != nil {
					log.Printf("json unmarshal body error, %s, %s, %s\n", err, apiKey, body)
				} else {
					varsMap["json_body_keys"] = b
				}
			}
		}
	}

	return a, varsMap
}

var templateRe = regexp.MustCompile(`<<.+>>`)

func decoder(in string, vars map[string]interface{}) string {
	if templateRe.FindStringIndex(in) != nil {
		var o bytes.Buffer
		tmpl, err := template.New("text").Delims("<<", ">>").Parse(in)
		if err != nil {
			log.Printf("decoder data parse error, in:%s, error:%v\n", in, err)
			return in
		}
		if err := tmpl.Execute(&o, vars); err != nil {
			log.Printf("decoder data execute error, in:%s, error:%v\n", in, err)
			return in
		}

		return o.String()
	}

	return in
}
