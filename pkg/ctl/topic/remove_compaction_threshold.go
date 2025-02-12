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

package topic

import (
	"github.com/streamnative/pulsarctl/pkg/cmdutils"

	util "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
)

func RemoveCompactionThresholdCmd(vc *cmdutils.VerbCmd) {
	var desc cmdutils.LongDescription
	desc.CommandUsedFor = "Remove the compaction threshold for a topic"
	desc.CommandPermission = "This command requires tenant admin permissions."

	var examples []cmdutils.Example
	set := cmdutils.Example{
		Desc:    "Remove the compaction threshold for a topic",
		Command: "pulsarctl topics remove-compaction-threshold topic",
	}
	examples = append(examples, set)
	desc.CommandExamples = examples

	var out []cmdutils.Output
	successOut := cmdutils.Output{
		Desc: "normal output",
		Out:  "Successfully remove compaction threshold for topic (topic-name)",
	}
	out = append(out, successOut)
	desc.CommandOutput = out

	vc.SetDescription(
		"remove-compaction-threshold",
		desc.CommandUsedFor,
		desc.ToString(),
		desc.ExampleToString())

	vc.SetRunFuncWithNameArg(func() error {
		return doRemoveCompactionThreshold(vc)
	}, "the topic name is not specified or the topic name is specified more than one")
}

func doRemoveCompactionThreshold(vc *cmdutils.VerbCmd) error {
	topic, err := util.GetTopicName(vc.NameArg)
	if err != nil {
		return err
	}

	admin := cmdutils.NewPulsarClient()
	err = admin.Topics().RemoveCompactionThreshold(*topic)
	if err == nil {
		vc.Command.Printf("Successfully remove compaction threshold for topic %s", topic)
	}

	return err
}
