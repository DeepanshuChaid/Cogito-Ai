package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/adapters"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/logger"
	"github.com/DeepanshuChaid/Cogito-Ai.git/pkg/types"
)

func main() {
	platform := flag.String("platform", "claude-code", "AI platform (claude-code, cursor, gemini-cli)")
	event := flag.String("event", "session_start", "Hook event type")
	configFile := flag.String("config", "", "Path to config file")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// LOAD CONFIG
	config := config.MustLoad(*configFile)
	config.Debug = *debug
	config.Platform = types.Platform(*platform)

	// INIT THE COMPONENTS
	adapter := adapters.GetAdapter(config.Platform)
	injector := injector.NewInjector(config)

	// READ STDIN (THE HOOK INPUT FROM THE AI CLI
	rawInput, err := io.ReadAll(os.Stdin)
	if err != nil {
		logger.LogFatal(config.Debug, "Failed TO parse Input %v", err)
	}


	// PARSE INPUT THROUGH ADAPTER
	input, err := adapter.ParseInput(rawInput)
	if err != nil {
		logger.LogFatal(config.Debug, "Failed to parse input: %v", input)
	}

	input.Event = types.HookEvent(*event)

	if config.Debug {
		log.Printf("[DEBUG] RECIEVED INPUT: %v", input)
	}

	// INJECT CONTEXT
	context, err := injector.InjectContext(input)
	if err != nil {
		logger.LogFatal(config.Debug, "Failed to inject context: %v", err)
	}

	output := &types.HookOutput{
		Continue: true,
		SupressOutput: true,
		AdditionContext: context,
	}

	if config.Debug && context != "" {
		output.SystemMessage = fmt.Sprintf("\n\n Cogito Context Injected (%d chars)", len(context))
	}

	// FORMAT AND PRINT OUTPUT
	formatted, err := adapter.FormatOuput(output)
	if err != nil {
		logger.LogFatal(config.Debug, "Failed to Format output: %v", err)
	}

	fmt.Println(string(formatted))
}
