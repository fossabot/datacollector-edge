/*
 * Copyright 2017 StreamSets Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package dev_random

import (
	"github.com/streamsets/datacollector-edge/api"
	"github.com/streamsets/datacollector-edge/api/fieldtype"
	"github.com/streamsets/datacollector-edge/container/common"
	"github.com/streamsets/datacollector-edge/container/creation"
	"github.com/streamsets/datacollector-edge/container/execution/runner"
	"testing"
)

const (
	CONF_RAW_DATA    = "rawData"
	CONF_DATA_FORMAT = "dataFormat"
)

func getStageContext(
	rawData string,
	dataFormat string,
	parameters map[string]interface{},
) *common.StageContextImpl {
	stageConfig := common.StageConfiguration{}
	stageConfig.Library = LIBRARY
	stageConfig.StageName = STAGE_NAME
	stageConfig.Configuration = []common.Config{
		{
			Name:  CONF_RAW_DATA,
			Value: rawData,
		},
		{
			Name:  CONF_DATA_FORMAT,
			Value: dataFormat,
		},
	}
	return &common.StageContextImpl{
		StageConfig: &stageConfig,
		Parameters:  parameters,
	}
}

func TestDevRawDataDSource_Init(t *testing.T) {
	rawData := "text1\n text2"
	stageContext := getStageContext(rawData, "TEXT", nil)
	stageBean, err := creation.NewStageBean(stageContext.StageConfig, stageContext.Parameters)
	if err != nil {
		t.Error(err)
	}
	stageInstance := stageBean.Stage

	if stageInstance.(*DevRawDataDSource).RawData != rawData {
		t.Error("Failed to inject config value for rawData")
	}
}

func TestDevRandomOrigin_TextFormat(t *testing.T) {
	rawData := "text1\n text2"
	stageContext := getStageContext(rawData, "TEXT", nil)

	stageBean, err := creation.NewStageBean(stageContext.StageConfig, stageContext.Parameters)
	if err != nil {
		t.Error(err)
	}
	stageInstance := stageBean.Stage

	issues := stageInstance.Init(stageContext)
	if len(issues) != 0 {
		t.Error(issues[0].Message)
	}

	batchMaker := runner.NewBatchMakerImpl(runner.StagePipe{})
	_, err = stageInstance.(api.Origin).Produce("", 5, batchMaker)
	if err != nil {
		t.Error("Err :", err)
	}

	records := batchMaker.GetStageOutput()
	if len(records) != 2 {
		t.Error("Excepted 2 records but got - ", len(records))
	}

	for _, record := range records {
		rootField, _ := record.Get()
		if rootField.Type != fieldtype.MAP {
			t.Error("Excepted Map field type but got - ", rootField.Type, " Value: ", rootField.Value)
			return
		}
	}
	stageInstance.Destroy()
}

func TestDevRandomOrigin_JsonFormat(t *testing.T) {
	rawData := "{\n  \"f1\": \"abc\"\n}\n{\n  \"f1\": \"xyz\"\n}"
	stageContext := getStageContext(rawData, "JSON", nil)

	stageBean, err := creation.NewStageBean(stageContext.StageConfig, stageContext.Parameters)
	if err != nil {
		t.Error(err)
	}
	stageInstance := stageBean.Stage

	issues := stageInstance.Init(stageContext)
	if len(issues) != 0 {
		t.Error(issues[0].Message)
	}

	batchMaker := runner.NewBatchMakerImpl(runner.StagePipe{})
	_, err = stageInstance.(api.Origin).Produce("", 5, batchMaker)
	if err != nil {
		t.Error("Err :", err)
	}

	records := batchMaker.GetStageOutput()
	if len(records) != 2 {
		t.Error("Excepted 2 records but got - ", len(records))
		return
	}

	record1 := records[0]
	rootField, _ := record1.Get()
	if rootField.Type != fieldtype.MAP {
		t.Error("Excepted Map field type but got - ", rootField.Type, " Value: ", rootField.Value)
		return
	}
	rootFieldValue := rootField.Value.(map[string]*api.Field)
	for key, field := range rootFieldValue {
		if field.Type != fieldtype.STRING {
			t.Error("Excepted String field type but got - ", field.Type, " Value: ", field.Value)
			return
		}
		if key != "f1" {
			t.Error("Invalid key", key)
		}
		if field.Value != "abc" {
			t.Error("Invalid value", field.Value)
		}
	}

	stageInstance.Destroy()
}
