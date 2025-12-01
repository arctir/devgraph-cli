// Package commands provides command-line command implementations for Devgraph CLI.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kong"
)

// CompletionCommand generates shell completion scripts for the Devgraph CLI.
type CompletionCommand struct {
	Shell   string `kong:"arg,optional,help='Shell type (bash, zsh, fish, powershell). Auto-detects if not specified.'"`
	Install bool   `kong:"help='Install completion script to the appropriate location'"`
}

// Run executes the completion command, generating shell completion scripts.
func (c *CompletionCommand) Run(ctx *kong.Context) error {
	// Auto-detect shell if not specified
	shell := c.Shell
	if shell == "" {
		shell = detectShell()
		if shell == "" {
			return fmt.Errorf("unable to detect shell type. Please specify one of: bash, zsh, fish, powershell")
		}
	}

	// Validate shell type
	validShells := map[string]bool{
		"bash":       true,
		"zsh":        true,
		"fish":       true,
		"powershell": true,
	}
	if !validShells[shell] {
		return fmt.Errorf("unsupported shell: %s. Supported shells: bash, zsh, fish, powershell", shell)
	}

	// Generate completion script
	script, err := generateCompletionScript(ctx, shell)
	if err != nil {
		return fmt.Errorf("failed to generate completion script: %w", err)
	}

	// Install or print
	if c.Install {
		return installCompletion(shell, script)
	}

	fmt.Print(script)
	return nil
}

// detectShell attempts to detect the current shell from the SHELL environment variable
func detectShell() string {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return ""
	}

	shellName := filepath.Base(shellPath)
	switch shellName {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "powershell", "pwsh":
		return "powershell"
	default:
		return ""
	}
}

// generateCompletionScript generates the completion script for the specified shell
func generateCompletionScript(ctx *kong.Context, shell string) (string, error) {
	switch shell {
	case "bash":
		return generateBashCompletion(ctx), nil
	case "zsh":
		return generateZshCompletion(ctx), nil
	case "fish":
		return generateFishCompletion(ctx), nil
	case "powershell":
		return generatePowershellCompletion(ctx), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// generateBashCompletion generates a bash completion script
func generateBashCompletion(ctx *kong.Context) string {
	commands := getCommands()
	return fmt.Sprintf(`# bash completion for %s

# Helper function to get dynamic completions
_%s_dynamic() {
    local resource_type="$1"
    %s complete "$resource_type" 2>/dev/null
}

_%s_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Top-level commands
    local commands="%s"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands} --help -h" -- ${cur}) )
        return 0
    fi

    # Subcommands for specific commands
    case "${COMP_WORDS[1]}" in
        auth)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "login logout whoami token --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        config)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "get-contexts current-context current-env use-context set-context delete-context get-clusters set-cluster delete-cluster get-users set-credentials delete-user --help" -- ${cur}) )
            else
                # Handle dynamic completions for config subcommands
                case "${COMP_WORDS[2]}" in
                    use-context|delete-context)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local contexts=$(_%s_dynamic contexts)
                            COMPREPLY=( $(compgen -W "${contexts}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        fi
                        ;;
                    set-context)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local contexts=$(_%s_dynamic contexts)
                            COMPREPLY=( $(compgen -W "${contexts}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--cluster --user --env --help" -- ${cur}) )
                        fi
                        ;;
                    delete-cluster)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local clusters=$(_%s_dynamic clusters)
                            COMPREPLY=( $(compgen -W "${clusters}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        fi
                        ;;
                    set-cluster)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local clusters=$(_%s_dynamic clusters)
                            COMPREPLY=( $(compgen -W "${clusters}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--server --issuer-url --client-id --help" -- ${cur}) )
                        fi
                        ;;
                    delete-user)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local users=$(_%s_dynamic users)
                            COMPREPLY=( $(compgen -W "${users}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        fi
                        ;;
                    set-credentials)
                        if [[ ${COMP_CWORD} -eq 3 ]]; then
                            local users=$(_%s_dynamic users)
                            COMPREPLY=( $(compgen -W "${users}" -- ${cur}) )
                        else
                            COMPREPLY=( $(compgen -W "--access-token --refresh-token --id-token --help" -- ${cur}) )
                        fi
                        ;;
                    get-contexts|get-clusters|get-users)
                        COMPREPLY=( $(compgen -W "--output -o --help" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            fi
            ;;
        env)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "current list --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        user)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list add remove --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        suggestion)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list create delete --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        entity-definition)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local defs=$(_%s_dynamic entity-definitions)
                        COMPREPLY=( $(compgen -W "${defs}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        entity)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local entities=$(_%s_dynamic entities)
                        COMPREPLY=( $(compgen -W "${entities}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        mcp)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local mcps=$(_%s_dynamic mcps)
                        COMPREPLY=( $(compgen -W "${mcps}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        modelprovider)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local providers=$(_%s_dynamic modelproviders)
                        COMPREPLY=( $(compgen -W "${providers}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        model)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local models=$(_%s_dynamic models)
                        COMPREPLY=( $(compgen -W "${models}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        oauthservice)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local services=$(_%s_dynamic oauthservices)
                        COMPREPLY=( $(compgen -W "${services}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        provider)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list get create update delete --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local providers=$(_%s_dynamic providers)
                        COMPREPLY=( $(compgen -W "${providers}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        subscription)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "list --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            fi
            ;;
        token)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "create delete get list update --help" -- ${cur}) )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                case "${COMP_WORDS[2]}" in
                    get|update|delete)
                        local tokens=$(_%s_dynamic tokens)
                        COMPREPLY=( $(compgen -W "${tokens}" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            else
                case "${COMP_WORDS[2]}" in
                    update)
                        COMPREPLY=( $(compgen -W "--name --scopes --help" -- ${cur}) )
                        ;;
                    *)
                        COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
                        ;;
                esac
            fi
            ;;
        completion)
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "bash zsh fish powershell --install --help" -- ${cur}) )
            else
                COMPREPLY=( $(compgen -W "--install --help" -- ${cur}) )
            fi
            ;;
        chat)
            COMPREPLY=( $(compgen -W "--help -h --model -m --max-tokens -t --stream -s --debug -d" -- ${cur}) )
            ;;
        *)
            COMPREPLY=( $(compgen -W "--help" -- ${cur}) )
            ;;
    esac

    return 0
}

complete -F _%s_completions %s
`, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, commands,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name)
}

// generateZshCompletion generates a zsh completion script
func generateZshCompletion(ctx *kong.Context) string {
	return fmt.Sprintf(`#compdef %s

# Helper function to get dynamic completions
_%s_dynamic() {
    local resource_type="$1"
    %s complete "$resource_type" 2>/dev/null
}

_%s() {
    local line state

    _arguments -C \
        "1: :_%s_commands" \
        "*::arg:->args"

    case $line[1] in
        auth)
            _arguments "1: :(login logout whoami token)"
            ;;
        config)
            case $line[2] in
                use-context|delete-context)
                    local contexts; contexts=(${(f)"$(_%s_dynamic contexts)"})
                    _arguments "1: :($contexts)"
                    ;;
                set-context)
                    local contexts; contexts=(${(f)"$(_%s_dynamic contexts)"})
                    _arguments "1: :($contexts)" "--cluster[Cluster name]" "--user[User name]" "--env[Environment ID]"
                    ;;
                delete-cluster|set-cluster)
                    local clusters; clusters=(${(f)"$(_%s_dynamic clusters)"})
                    _arguments "1: :($clusters)"
                    ;;
                delete-user|set-credentials)
                    local users; users=(${(f)"$(_%s_dynamic users)"})
                    _arguments "1: :($users)"
                    ;;
                *)
                    _arguments "1: :(get-contexts current-context current-env use-context set-context delete-context get-clusters set-cluster delete-cluster get-users set-credentials delete-user)"
                    ;;
            esac
            ;;
        env)
            _arguments "1: :(current list)"
            ;;
        user)
            _arguments "1: :(list add remove)"
            ;;
        suggestion)
            _arguments "1: :(list create delete)"
            ;;
        entity-definition)
            case $line[2] in
                get|update|delete)
                    local defs; defs=(${(f)"$(_%s_dynamic entity-definitions)"})
                    _arguments "1: :($defs)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        entity)
            case $line[2] in
                get|update|delete)
                    local entities; entities=(${(f)"$(_%s_dynamic entities)"})
                    _arguments "1: :($entities)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        mcp)
            case $line[2] in
                get|update|delete)
                    local mcps; mcps=(${(f)"$(_%s_dynamic mcps)"})
                    _arguments "1: :($mcps)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        modelprovider)
            case $line[2] in
                get|update|delete)
                    local providers; providers=(${(f)"$(_%s_dynamic modelproviders)"})
                    _arguments "1: :($providers)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        model)
            case $line[2] in
                get|update|delete)
                    local models; models=(${(f)"$(_%s_dynamic models)"})
                    _arguments "1: :($models)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        oauthservice)
            case $line[2] in
                get|update|delete)
                    local services; services=(${(f)"$(_%s_dynamic oauthservices)"})
                    _arguments "1: :($services)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        provider)
            case $line[2] in
                get|update|delete)
                    local providers; providers=(${(f)"$(_%s_dynamic providers)"})
                    _arguments "1: :($providers)"
                    ;;
                *)
                    _arguments "1: :(list get create update delete)"
                    ;;
            esac
            ;;
        token)
            case $line[2] in
                get|update|delete)
                    local tokens; tokens=(${(f)"$(_%s_dynamic tokens)"})
                    _arguments "1: :($tokens)"
                    ;;
                *)
                    _arguments "1: :(create delete get list update)"
                    ;;
            esac
            ;;
        subscription)
            _arguments "1: :(list)"
            ;;
        completion)
            _arguments "1: :(bash zsh fish powershell)" "--install[Install completion script]"
            ;;
    esac
}

_%s_commands() {
    local commands; commands=(
        %s
    )
    _describe -t commands '%s commands' commands
}

_%s "$@"
`, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		getCommandsWithDescriptions(), ctx.Model.Name, ctx.Model.Name)
}

// generateFishCompletion generates a fish completion script
func generateFishCompletion(ctx *kong.Context) string {
	return fmt.Sprintf(`# fish completion for %s

# Remove default completions
complete -c %s -e

# Helper function for dynamic completions
function __%s_dynamic
    %s complete $argv[1] 2>/dev/null
end

# Top-level commands
complete -c %s -f -n "__fish_use_subcommand" -a "chat" -d "Start an interactive chat with AI"
complete -c %s -f -n "__fish_use_subcommand" -a "auth" -d "Manage authentication"
complete -c %s -f -n "__fish_use_subcommand" -a "config" -d "Manage configuration settings"
complete -c %s -f -n "__fish_use_subcommand" -a "token" -d "Manage opaque tokens"
complete -c %s -f -n "__fish_use_subcommand" -a "env" -d "Manage environments"
complete -c %s -f -n "__fish_use_subcommand" -a "entity-definition" -d "Manage entity definitions"
complete -c %s -f -n "__fish_use_subcommand" -a "entity" -d "Manage entities"
complete -c %s -f -n "__fish_use_subcommand" -a "mcp" -d "Manage MCP resources"
complete -c %s -f -n "__fish_use_subcommand" -a "modelprovider" -d "Manage Model Provider resources"
complete -c %s -f -n "__fish_use_subcommand" -a "model" -d "Manage Model resources"
complete -c %s -f -n "__fish_use_subcommand" -a "oauthservice" -d "Manage OAuth services"
complete -c %s -f -n "__fish_use_subcommand" -a "subscription" -d "Manage subscriptions"
complete -c %s -f -n "__fish_use_subcommand" -a "suggestion" -d "Manage chat suggestions"
complete -c %s -f -n "__fish_use_subcommand" -a "provider" -d "Manage discovery providers"
complete -c %s -f -n "__fish_use_subcommand" -a "user" -d "Manage users in the current environment"
complete -c %s -f -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion scripts"

# Auth subcommands
complete -c %s -f -n "__fish_seen_subcommand_from auth" -a "login" -d "Authenticate with your account"
complete -c %s -f -n "__fish_seen_subcommand_from auth" -a "logout" -d "Log out and clear credentials"
complete -c %s -f -n "__fish_seen_subcommand_from auth" -a "whoami" -d "Show current user info"
complete -c %s -f -n "__fish_seen_subcommand_from auth" -a "token" -d "Print authentication token"

# Config subcommands
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "get-contexts" -d "List all contexts"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "current-context" -d "Display the current context"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "current-env" -d "Display the current environment ID"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "use-context" -d "Set the current context"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "set-context" -d "Create or modify a context"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "delete-context" -d "Delete a context"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "get-clusters" -d "List all clusters"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "set-cluster" -d "Create or modify a cluster"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "delete-cluster" -d "Delete a cluster"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "get-users" -d "List all users"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "set-credentials" -d "Set user credentials"
complete -c %s -f -n "__fish_seen_subcommand_from config" -a "delete-user" -d "Delete a user"

# Dynamic completions for config subcommands
complete -c %s -f -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from use-context delete-context set-context" -a "(__%s_dynamic contexts)"
complete -c %s -f -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from delete-cluster set-cluster" -a "(__%s_dynamic clusters)"
complete -c %s -f -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from delete-user set-credentials" -a "(__%s_dynamic users)"

# Common CRUD subcommands
set -l crud_commands "entity-definition entity mcp modelprovider model oauthservice provider"
complete -c %s -f -n "__fish_seen_subcommand_from $crud_commands" -a "list" -d "List resources"
complete -c %s -f -n "__fish_seen_subcommand_from $crud_commands" -a "get" -d "Get resource details"
complete -c %s -f -n "__fish_seen_subcommand_from $crud_commands" -a "create" -d "Create resource"
complete -c %s -f -n "__fish_seen_subcommand_from $crud_commands" -a "update" -d "Update resource"
complete -c %s -f -n "__fish_seen_subcommand_from $crud_commands" -a "delete" -d "Delete resource"

# Dynamic completions for CRUD resources
complete -c %s -f -n "__fish_seen_subcommand_from entity-definition; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic entity-definitions)"
complete -c %s -f -n "__fish_seen_subcommand_from entity; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic entities)"
complete -c %s -f -n "__fish_seen_subcommand_from mcp; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic mcps)"
complete -c %s -f -n "__fish_seen_subcommand_from modelprovider; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic modelproviders)"
complete -c %s -f -n "__fish_seen_subcommand_from model; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic models)"
complete -c %s -f -n "__fish_seen_subcommand_from oauthservice; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic oauthservices)"
complete -c %s -f -n "__fish_seen_subcommand_from provider; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic providers)"

# Token subcommands
complete -c %s -f -n "__fish_seen_subcommand_from token" -a "create" -d "Create token"
complete -c %s -f -n "__fish_seen_subcommand_from token" -a "delete" -d "Delete token"
complete -c %s -f -n "__fish_seen_subcommand_from token" -a "get" -d "Get token by ID"
complete -c %s -f -n "__fish_seen_subcommand_from token" -a "list" -d "List tokens"
complete -c %s -f -n "__fish_seen_subcommand_from token" -a "update" -d "Update token"
complete -c %s -f -n "__fish_seen_subcommand_from token; and __fish_seen_subcommand_from get update delete" -a "(__%s_dynamic tokens)"

# Env subcommands
complete -c %s -f -n "__fish_seen_subcommand_from env" -a "current" -d "Display current environment"
complete -c %s -f -n "__fish_seen_subcommand_from env" -a "list" -d "List environments"

# User subcommands
complete -c %s -f -n "__fish_seen_subcommand_from user" -a "list" -d "List users"
complete -c %s -f -n "__fish_seen_subcommand_from user" -a "add" -d "Invite a user"
complete -c %s -f -n "__fish_seen_subcommand_from user" -a "remove" -d "Remove a user"

# Suggestion subcommands
complete -c %s -f -n "__fish_seen_subcommand_from suggestion" -a "list" -d "List chat suggestions"
complete -c %s -f -n "__fish_seen_subcommand_from suggestion" -a "create" -d "Create a chat suggestion"
complete -c %s -f -n "__fish_seen_subcommand_from suggestion" -a "delete" -d "Delete a chat suggestion"

# Subscription subcommands
complete -c %s -f -n "__fish_seen_subcommand_from subscription" -a "list" -d "List subscriptions"

# Completion subcommands
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "bash" -d "Generate bash completion"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "zsh" -d "Generate zsh completion"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "fish" -d "Generate fish completion"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -a "powershell" -d "Generate powershell completion"
complete -c %s -f -n "__fish_seen_subcommand_from completion" -l "install" -d "Install completion script"
`, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name)
}

// generatePowershellCompletion generates a PowerShell completion script
func generatePowershellCompletion(ctx *kong.Context) string {
	return fmt.Sprintf(`# PowerShell completion for %s

# Helper function for dynamic completions
function Get-%sDynamic {
    param($ResourceType)
    $result = & %s complete $ResourceType 2>$null
    if ($result) {
        return $result -split [char]10 | Where-Object { $_ -ne '' }
    }
    return @()
}

Register-ArgumentCompleter -Native -CommandName '%s' -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $commandElements = $commandAst.CommandElements
    $command = @(
        '%s'
        for ($i = 1; $i -lt $commandElements.Count; $i++) {
            $element = $commandElements[$i]
            if ($element -isnot [System.Management.Automation.Language.StringConstantExpressionAst]) {
                break
            }
            $element.Value
        }
    )

    $completions = @()

    switch ($command.Count) {
        1 {
            $completions = @(
                @{Text='chat'; Description='Start an interactive chat with AI'},
                @{Text='auth'; Description='Manage authentication'},
                @{Text='config'; Description='Manage configuration settings'},
                @{Text='token'; Description='Manage opaque tokens'},
                @{Text='env'; Description='Manage environments'},
                @{Text='entity-definition'; Description='Manage entity definitions'},
                @{Text='entity'; Description='Manage entities'},
                @{Text='mcp'; Description='Manage MCP resources'},
                @{Text='modelprovider'; Description='Manage Model Provider resources'},
                @{Text='model'; Description='Manage Model resources'},
                @{Text='oauthservice'; Description='Manage OAuth services'},
                @{Text='subscription'; Description='Manage subscriptions'},
                @{Text='suggestion'; Description='Manage chat suggestions'},
                @{Text='provider'; Description='Manage discovery providers'},
                @{Text='user'; Description='Manage users in the current environment'},
                @{Text='completion'; Description='Generate shell completion scripts'}
            )
        }
        2 {
            switch ($command[1]) {
                'auth' {
                    $completions = @(
                        @{Text='login'; Description='Authenticate with your account'},
                        @{Text='logout'; Description='Log out and clear credentials'},
                        @{Text='whoami'; Description='Show current user info'},
                        @{Text='token'; Description='Print authentication token'}
                    )
                }
                'config' {
                    $completions = @(
                        @{Text='get-contexts'; Description='List all contexts'},
                        @{Text='current-context'; Description='Display the current context'},
                        @{Text='current-env'; Description='Display the current environment ID'},
                        @{Text='use-context'; Description='Set the current context'},
                        @{Text='set-context'; Description='Create or modify a context'},
                        @{Text='delete-context'; Description='Delete a context'},
                        @{Text='get-clusters'; Description='List all clusters'},
                        @{Text='set-cluster'; Description='Create or modify a cluster'},
                        @{Text='delete-cluster'; Description='Delete a cluster'},
                        @{Text='get-users'; Description='List all users'},
                        @{Text='set-credentials'; Description='Set user credentials'},
                        @{Text='delete-user'; Description='Delete a user'}
                    )
                }
                'env' {
                    $completions = @(
                        @{Text='current'; Description='Display current environment'},
                        @{Text='list'; Description='List environments'}
                    )
                }
                'user' {
                    $completions = @(
                        @{Text='list'; Description='List users'},
                        @{Text='add'; Description='Invite a user'},
                        @{Text='remove'; Description='Remove a user'}
                    )
                }
                'suggestion' {
                    $completions = @(
                        @{Text='list'; Description='List chat suggestions'},
                        @{Text='create'; Description='Create a chat suggestion'},
                        @{Text='delete'; Description='Delete a chat suggestion'}
                    )
                }
                {$_ -in @('entity-definition','entity','mcp','modelprovider','model','oauthservice','provider')} {
                    $completions = @(
                        @{Text='list'; Description='List resources'},
                        @{Text='get'; Description='Get resource details'},
                        @{Text='create'; Description='Create resource'},
                        @{Text='update'; Description='Update resource'},
                        @{Text='delete'; Description='Delete resource'}
                    )
                }
                'token' {
                    $completions = @(
                        @{Text='create'; Description='Create token'},
                        @{Text='delete'; Description='Delete token'},
                        @{Text='get'; Description='Get token by ID'},
                        @{Text='list'; Description='List tokens'},
                        @{Text='update'; Description='Update token'}
                    )
                }
                'subscription' {
                    $completions = @(
                        @{Text='list'; Description='List subscriptions'}
                    )
                }
                'completion' {
                    $completions = @(
                        @{Text='bash'; Description='Generate bash completion'},
                        @{Text='zsh'; Description='Generate zsh completion'},
                        @{Text='fish'; Description='Generate fish completion'},
                        @{Text='powershell'; Description='Generate powershell completion'}
                    )
                }
            }
        }
        3 {
            # Dynamic completions for third argument
            switch ($command[1]) {
                'config' {
                    switch ($command[2]) {
                        {$_ -in @('use-context', 'delete-context', 'set-context')} {
                            $contexts = Get-%sDynamic 'contexts'
                            $completions = $contexts | ForEach-Object { @{Text=$_; Description='Context'} }
                        }
                        {$_ -in @('delete-cluster', 'set-cluster')} {
                            $clusters = Get-%sDynamic 'clusters'
                            $completions = $clusters | ForEach-Object { @{Text=$_; Description='Cluster'} }
                        }
                        {$_ -in @('delete-user', 'set-credentials')} {
                            $users = Get-%sDynamic 'users'
                            $completions = $users | ForEach-Object { @{Text=$_; Description='User'} }
                        }
                    }
                }
                'entity-definition' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $defs = Get-%sDynamic 'entity-definitions'
                        $completions = $defs | ForEach-Object { @{Text=$_; Description='Entity Definition'} }
                    }
                }
                'entity' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $entities = Get-%sDynamic 'entities'
                        $completions = $entities | ForEach-Object { @{Text=$_; Description='Entity'} }
                    }
                }
                'mcp' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $mcps = Get-%sDynamic 'mcps'
                        $completions = $mcps | ForEach-Object { @{Text=$_; Description='MCP'} }
                    }
                }
                'modelprovider' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $providers = Get-%sDynamic 'modelproviders'
                        $completions = $providers | ForEach-Object { @{Text=$_; Description='Model Provider'} }
                    }
                }
                'model' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $models = Get-%sDynamic 'models'
                        $completions = $models | ForEach-Object { @{Text=$_; Description='Model'} }
                    }
                }
                'oauthservice' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $services = Get-%sDynamic 'oauthservices'
                        $completions = $services | ForEach-Object { @{Text=$_; Description='OAuth Service'} }
                    }
                }
                'provider' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $providers = Get-%sDynamic 'providers'
                        $completions = $providers | ForEach-Object { @{Text=$_; Description='Provider'} }
                    }
                }
                'token' {
                    if ($command[2] -in @('get', 'update', 'delete')) {
                        $tokens = Get-%sDynamic 'tokens'
                        $completions = $tokens | ForEach-Object { @{Text=$_; Description='Token'} }
                    }
                }
            }
        }
    }

    $completions | Where-Object { $_.Text -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_.Text, $_.Text, 'ParameterValue', $_.Description)
    }
}
`, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name, ctx.Model.Name,
		ctx.Model.Name)
}

// installCompletion installs the completion script to the appropriate location
func installCompletion(shell, script string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var path string
	var instructions string

	switch shell {
	case "bash":
		if runtime.GOOS == "darwin" {
			// macOS with Homebrew bash-completion
			path = "/usr/local/etc/bash_completion.d/dg"
			instructions = "Ensure bash-completion is installed via Homebrew and loaded in your ~/.bash_profile"
		} else {
			// Linux
			path = filepath.Join(homeDir, ".local/share/bash-completion/completions/dg")
			instructions = "Add this line to your ~/.bashrc if not present:\n  [[ -d ~/.local/share/bash-completion/completions ]] && for f in ~/.local/share/bash-completion/completions/*; do source \"$f\"; done"
		}

	case "zsh":
		path = filepath.Join(homeDir, ".zsh/completions/_dg")
		instructions = "Add this line to your ~/.zshrc if not present:\n  fpath=(~/.zsh/completions $fpath) && autoload -Uz compinit && compinit"

	case "fish":
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(homeDir, ".config")
		}
		path = filepath.Join(configDir, "fish/completions/dg.fish")
		instructions = "Fish will automatically load completions from this location"

	case "powershell":
		if runtime.GOOS == "windows" {
			path = filepath.Join(homeDir, "Documents/PowerShell/Scripts/dg-completion.ps1")
		} else {
			path = filepath.Join(homeDir, ".config/powershell/Scripts/dg-completion.ps1")
		}
		instructions = "Add this line to your PowerShell profile:\n  . " + path

	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the completion script
	if err := os.WriteFile(path, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write completion script to %s: %w", path, err)
	}

	fmt.Printf("✅ Completion script installed to: %s\n\n", path)
	fmt.Printf("%s\n\n", instructions)
	fmt.Println("Restart your shell or source the completion file to enable completions.")

	return nil
}

// getCommands returns a space-separated list of top-level commands
func getCommands() string {
	return "chat auth config token env entity-definition entity mcp modelprovider model oauthservice subscription suggestion provider user completion"
}

// getCommandsWithDescriptions returns command list formatted for zsh completion with descriptions
func getCommandsWithDescriptions() string {
	return `'chat:Start an interactive chat with AI'
        'auth:Manage authentication'
        'config:Manage configuration settings'
        'token:Manage opaque tokens'
        'env:Manage environments'
        'entity-definition:Manage entity definitions'
        'entity:Manage entities'
        'mcp:Manage MCP resources'
        'modelprovider:Manage Model Provider resources'
        'model:Manage Model resources'
        'oauthservice:Manage OAuth services'
        'subscription:Manage subscriptions'
        'suggestion:Manage chat suggestions'
        'provider:Manage discovery providers'
        'user:Manage users in the current environment'
        'completion:Generate shell completion scripts'`
}
