package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/streamsets/datacollector-edge/container/common"
	"github.com/streamsets/datacollector-edge/container/execution/manager"
	"github.com/streamsets/datacollector-edge/container/process"
	"github.com/streamsets/datacollector-edge/container/store"
	"github.com/streamsets/datacollector-edge/container/util"
	"log"
	"net/http"
	"net/http/pprof"
)

type WebServerTask struct {
	config            Config
	buildInfo         *common.BuildInfo
	manager           manager.Manager
	pipelineStoreTask store.PipelineStoreTask
	httpServer        *http.Server
	processManager    *process.Manager
}

func (webServerTask *WebServerTask) Init() error {
	fmt.Println("Running on URI : http://localhost" + webServerTask.config.BindAddress)
	log.Println("[INFO] Running on URI : http://localhost" + webServerTask.config.BindAddress)

	router := httprouter.New()
	router.GET("/", webServerTask.homeHandler)

	// Manager APIs
	router.POST("/rest/v1/pipeline/:pipelineId/start", webServerTask.startHandler)
	router.POST("/rest/v1/pipeline/:pipelineId/stop", webServerTask.stopHandler)
	router.POST("/rest/v1/pipeline/:pipelineId/resetOffset", webServerTask.resetOffsetHandler)
	router.POST("/rest/v1/pipeline/:pipelineId/committedOffsets", webServerTask.updateOffsetHandler)

	router.GET("/rest/v1/pipeline/:pipelineId/status", webServerTask.statusHandler)
	router.GET("/rest/v1/pipeline/:pipelineId/history", webServerTask.historyHandler)
	router.GET("/rest/v1/pipeline/:pipelineId/metrics", webServerTask.metricsHandler)
	router.GET("/rest/v1/pipeline/:pipelineId/committedOffsets", webServerTask.getOffsetHandler)

	// Pipeline Store APIs
	router.GET("/rest/v1/pipelines", webServerTask.getPipelines)
	router.GET("/rest/v1/pipeline/:pipelineId", webServerTask.getPipeline)
	router.PUT("/rest/v1/pipeline/:pipelineTitle", webServerTask.createPipeline)
	router.POST("/rest/v1/pipeline/:pipelineId", webServerTask.savePipeline)

	// Register pprof handlers
	router.HandlerFunc("GET", "/debug/pprof/", pprof.Index)
	router.Handler("GET", "/debug/pprof/heap", pprof.Handler("heap"))
	router.Handler("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handler("GET", "/debug/pprof/block", pprof.Handler("block"))
	router.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
	router.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
	router.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
	router.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)

	router.GET("/rest/v1/processMetrics", webServerTask.processMetricsHandler)

	webServerTask.httpServer = &http.Server{Addr: webServerTask.config.BindAddress, Handler: router}
	return nil
}

func (webServerTask *WebServerTask) homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	encoder.Encode(webServerTask.buildInfo)
}

func (webServerTask *WebServerTask) Run() {
	fmt.Println(webServerTask.httpServer.ListenAndServe())
}

func (webServerTask *WebServerTask) Shutdown() {
	err := webServerTask.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Printf("[ERROR] Error happened when shutting webserver : %s\n", err.Error())
	}
}

func (webServerTask *WebServerTask) processMetricsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")
	encoder.Encode(util.FormatMetricsRegistry(webServerTask.processManager.GetProcessMetrics()))
}

func NewWebServerTask(
	config Config,
	buildInfo *common.BuildInfo,
	manager manager.Manager,
	pipelineStoreTask store.PipelineStoreTask,
	processManager	  *process.Manager,
) (*WebServerTask, error) {
	webServerTask := WebServerTask{
		config:                 config,
		buildInfo:              buildInfo,
		manager:                manager,
		pipelineStoreTask:      pipelineStoreTask,
		processManager:         processManager,
	}
	err := webServerTask.Init()
	if err != nil {
		return nil, err
	}
	return &webServerTask, nil
}
