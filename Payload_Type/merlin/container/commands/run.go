/*
Merlin is a post-exploitation command and control framework.

This file is part of Merlin.
Copyright (C) 2023  Russel Van Tuyl

Merlin is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as
published by the Free Software Foundation, either version 3 of the License, or any later version.

Merlin is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with Merlin.  If not, see <http://www.gnu.org/licenses/>.
*/

package commands

import (
	// Standard
	"fmt"
	"strings"

	// Mythic
	structs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"

	// Merlin Message
	"github.com/Ne0nd0g/merlin-message/jobs"
)

// run creates and return a Mythic Command structure that is registered with the Mythic server
func run() structs.Command {
	attr := structs.CommandAttribute{
		SupportedOS: []string{structs.SUPPORTED_OS_WINDOWS, structs.SUPPORTED_OS_LINUX, structs.SUPPORTED_OS_MACOS, "freebsd", "openbsd", "solaris"},
	}

	executable := structs.CommandParameter{
		Name:                                    "executable",
		ModalDisplayName:                        "Executable",
		CLIName:                                 "executable",
		ParameterType:                           structs.COMMAND_PARAMETER_TYPE_STRING,
		Description:                             "The executable, or program, to run",
		Choices:                                 nil,
		DefaultValue:                            nil,
		SupportedAgents:                         nil,
		SupportedAgentBuildParameters:           nil,
		ChoicesAreAllCommands:                   false,
		ChoicesAreLoadedCommands:                false,
		FilterCommandChoicesByCommandAttributes: nil,
		DynamicQueryFunction:                    nil,
		ParameterGroupInformation: []structs.ParameterGroupInfo{
			{
				ParameterIsRequired:   true,
				GroupName:             "Default",
				UIModalPosition:       0,
				AdditionalInformation: nil,
			},
		},
	}

	args := structs.CommandParameter{
		Name:                                    "arguments",
		ModalDisplayName:                        "Arguments",
		CLIName:                                 "args",
		ParameterType:                           structs.COMMAND_PARAMETER_TYPE_STRING,
		Description:                             "Arguments to start the executable with",
		Choices:                                 nil,
		DefaultValue:                            "",
		SupportedAgents:                         nil,
		SupportedAgentBuildParameters:           nil,
		ChoicesAreAllCommands:                   false,
		ChoicesAreLoadedCommands:                false,
		FilterCommandChoicesByCommandAttributes: nil,
		DynamicQueryFunction:                    nil,
		ParameterGroupInformation: []structs.ParameterGroupInfo{
			{
				ParameterIsRequired:   false,
				GroupName:             "Default",
				UIModalPosition:       1,
				AdditionalInformation: nil,
			},
		},
	}

	params := []structs.CommandParameter{executable, args}
	command := structs.Command{
		Name:                           "run",
		NeedsAdminPermissions:          false,
		HelpString:                     "run <arg1> <arg2>...",
		Description:                    "Run the executable with the provided arguments and return the results",
		Version:                        0,
		SupportedUIFeatures:            nil,
		Author:                         "@Ne0nd0g",
		MitreAttackMappings:            []string{"T1106"},
		ScriptOnlyCommand:              false,
		CommandAttributes:              attr,
		CommandParameters:              params,
		AssociatedBrowserScript:        nil,
		TaskFunctionOPSECPre:           nil,
		TaskFunctionCreateTasking:      runCreateTask,
		TaskFunctionProcessResponse:    nil,
		TaskFunctionOPSECPost:          nil,
		TaskFunctionParseArgString:     taskFunctionParseArgString,
		TaskFunctionParseArgDictionary: taskFunctionParseArgDictionary,
		TaskCompletionFunctions:        nil,
	}

	return command
}

// runCreateTask takes a Mythic Task and converts into a Merlin Job that is encoded into JSON and subsequently sent to the Merlin Agent
func runCreateTask(task *structs.PTTaskMessageAllData) (resp structs.PTTaskCreateTaskingMessageResponse) {
	pkg := "mythic/container/commands/run/runCreateTask()"
	resp.TaskID = task.Task.ID

	executable, err := task.Args.GetStringArg("executable")
	if err != nil {
		err = fmt.Errorf("%s: there was an error getting the 'executable' argument: %s", pkg, err)
		resp.Error = err.Error()
		resp.Success = false
		logging.LogError(err, "returning with error")
		return
	}

	args, err := task.Args.GetStringArg("arguments")
	if err != nil {
		err = fmt.Errorf("%s: there was an error getting the 'arguments' argument: %s", pkg, err)
		resp.Error = err.Error()
		resp.Success = false
		logging.LogError(err, "returning with error")
		return
	}

	job := jobs.Command{
		Command: executable,
	}

	// TODO parse args according to target OS (e.g., escapes, etc.)
	if args != "" {
		job.Args = append(job.Args, strings.Split(args, " ")...)
	}

	mythicJob, err := ConvertMerlinJobToMythicTask(job, jobs.CMD)
	if err != nil {
		err = fmt.Errorf("%s: there was an error converting the Merlin job to a Mythic job: %s", pkg, err)
		resp.Error = err.Error()
		resp.Success = false
		logging.LogError(err, "returning with error")
		return
	}

	task.Args.SetManualArgs(mythicJob)

	disp := fmt.Sprintf("%s %s", executable, args)
	resp.DisplayParams = &disp
	resp.Success = true

	return
}
