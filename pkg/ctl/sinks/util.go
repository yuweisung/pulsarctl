// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sinks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/streamnative/pulsarctl/pkg/cmdutils"
	"github.com/streamnative/pulsarctl/pkg/ctl/utils"
	"github.com/streamnative/pulsarctl/pkg/pulsar/common"
	util "github.com/streamnative/pulsarctl/pkg/pulsar/utils"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func processArguments(sinkData *util.SinkData) error {
	// Initialize config builder either from a supplied YAML config file or from scratch
	if sinkData.SinkConf != nil {
		// no-op
	} else {
		sinkData.SinkConf = new(util.SinkConfig)
	}

	if sinkData.SinkConfigFile != "" {
		yamlFile, err := ioutil.ReadFile(sinkData.SinkConfigFile)
		if err == nil {
			err = yaml.Unmarshal(yamlFile, sinkData.SinkConf)
			if err != nil {
				return fmt.Errorf("unmarshal yaml file error:%s", err.Error())
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("load conf file failed, err:%s", err.Error())
		}
	}

	if sinkData.Tenant != "" {
		sinkData.SinkConf.Tenant = sinkData.Tenant
	}

	if sinkData.Namespace != "" {
		sinkData.SinkConf.Namespace = sinkData.Namespace
	}

	if sinkData.Name != "" {
		sinkData.SinkConf.Name = sinkData.Name
	}

	if sinkData.ClassName != "" {
		sinkData.SinkConf.ClassName = sinkData.ClassName
	}

	if sinkData.ProcessingGuarantees != "" {
		sinkData.SinkConf.ProcessingGuarantees = sinkData.ProcessingGuarantees
	}

	if sinkData.RetainOrdering {
		sinkData.SinkConf.RetainOrdering = sinkData.RetainOrdering
	}

	if sinkData.Inputs != "" {
		inputTopics := strings.Split(sinkData.Inputs, ",")
		sinkData.SinkConf.Inputs = inputTopics
	}

	if sinkData.CustomSerdeInputString != "" {
		customSerdeInputMap := make(map[string]string)
		err := json.Unmarshal([]byte(sinkData.CustomSerdeInputString), &customSerdeInputMap)
		if err != nil {
			return err
		}
		sinkData.SinkConf.TopicToSerdeClassName = customSerdeInputMap
	}

	if sinkData.CustomSchemaInputString != "" {
		customSchemaInputMap := make(map[string]string)
		err := json.Unmarshal([]byte(sinkData.CustomSchemaInputString), &customSchemaInputMap)
		if err != nil {
			return err
		}
		sinkData.SinkConf.TopicToSchemaType = customSchemaInputMap
	}

	if sinkData.SubsName != "" {
		sinkData.SinkConf.SourceSubscriptionName = sinkData.SubsName
	}

	if sinkData.SubsPosition != "" {
		sinkData.SinkConf.SourceSubscriptionPosition = sinkData.SubsPosition
	}

	if sinkData.TopicsPattern != "" {
		sinkData.SinkConf.TopicsPattern = &sinkData.TopicsPattern
	}

	if sinkData.Parallelism != 0 {
		sinkData.SinkConf.Parallelism = sinkData.Parallelism
	} else {
		sinkData.SinkConf.Parallelism = 1
	}

	if sinkData.Archive != "" && sinkData.SinkType != "" {
		return errors.New("Cannot specify both archive and sink-type")
	}

	if sinkData.Archive != "" {
		sinkData.SinkConf.Archive = sinkData.Archive
	}

	if sinkData.SinkType != "" {
		sinkData.SinkConf.Archive = validateSinkType(sinkData.SinkType)
	}

	if sinkData.CPU != 0 {
		if sinkData.SinkConf.Resources == nil {
			sinkData.SinkConf.Resources = util.NewDefaultResources()
		}

		sinkData.SinkConf.Resources.CPU = sinkData.CPU
	}

	if sinkData.Disk != 0 {
		if sinkData.SinkConf.Resources == nil {
			sinkData.SinkConf.Resources = util.NewDefaultResources()
		}

		sinkData.SinkConf.Resources.Disk = sinkData.Disk
	}

	if sinkData.RAM != 0 {
		if sinkData.SinkConf.Resources == nil {
			sinkData.SinkConf.Resources = util.NewDefaultResources()
		}

		sinkData.SinkConf.Resources.RAM = sinkData.RAM
	}

	if sinkData.SinkConfigString != "" {
		sinkData.SinkConf.Configs = parseConfigs(sinkData.SinkConfigString)
	}

	if sinkData.AutoAck {
		sinkData.SinkConf.AutoAck = sinkData.AutoAck
	}

	if sinkData.TimeoutMs != 0 {
		sinkData.SinkConf.TimeoutMs = &sinkData.TimeoutMs
	}

	return nil
}

func validateSinkType(sinkType string) string {
	availableSinks := make([]string, 0, 10)
	admin := cmdutils.NewPulsarClientWithAPIVersion(common.V3)
	connectorDefinition, err := admin.Sinks().GetBuiltInSinks()
	if err != nil {
		log.Printf("get builtin sinks error: %s\n", err.Error())
		return ""
	}

	for _, value := range connectorDefinition {
		availableSinks = append(availableSinks, value.Name)
	}

	availableSinksString := strings.Join(availableSinks, " ")
	if !strings.Contains(availableSinksString, sinkType) {
		log.Printf("invalid sink type [%s] -- Available sinks are: %s", sinkType, availableSinks)
		return ""
	}

	// Sink type is a valid built-in connector type
	return "builtin://" + sinkType
}

func parseConfigs(str string) map[string]interface{} {
	var resMap map[string]interface{}

	err := json.Unmarshal([]byte(str), &resMap)
	if err != nil {
		return nil
	}

	return resMap
}

func validateSinkConfigs(sinkConf *util.SinkConfig) error {
	if sinkConf.Archive == "" {
		return errors.New("Sink archive not specified")
	}

	utils.InferMissingSinkeArguments(sinkConf)

	if utils.IsPackageURLSupported(sinkConf.Archive) && strings.HasPrefix(sinkConf.Archive, utils.BUILTIN) {
		if !utils.IsFileExist(sinkConf.Archive) {
			return fmt.Errorf("sink Archive %s does not exist", sinkConf.Archive)
		}
	}

	if sinkConf.Name == "" {
		return errors.New("sink name not specified")
	}
	return nil
}

func checkArgsForUpdate(sinkConf *util.SinkConfig) {
	if sinkConf.Tenant == "" {
		sinkConf.Tenant = utils.PublicTenant
	}

	if sinkConf.Namespace == "" {
		sinkConf.Namespace = utils.DefaultNamespace
	}
}

func processNamespaceCmd(sinkData *util.SinkData) {
	if sinkData.Tenant == "" || sinkData.Namespace == "" {
		sinkData.Tenant = utils.PublicTenant
		sinkData.Namespace = utils.DefaultNamespace
	}
}

func processBaseArguments(sinkData *util.SinkData) error {
	processNamespaceCmd(sinkData)

	if sinkData.Name == "" {
		return errors.New("You must specify a name for the sink")
	}

	return nil
}
