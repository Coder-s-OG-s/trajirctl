# Print a ready to paste mcpServers block after you build trajir-mcp.
# Usage:
#   .\scripts\install-mcp-example.ps1 -Bin C:\path\to\trajir-mcp.exe -Root C:\path\to\project
param(
    [Parameter(Mandatory = $true)][string]$Bin,
    [Parameter(Mandatory = $true)][string]$Root
)

if (-not [System.IO.Path]::IsPathRooted($Bin) -or -not [System.IO.Path]::IsPathRooted($Root)) {
    Write-Error "Both -Bin and -Root must be absolute paths"
    exit 2
}

@{
    mcpServers = @{
        "trajectory-ir" = @{
            command = $Bin
            args    = @()
            env     = @{
                TRAJIR_MCP_ROOT = $Root
            }
        }
    }
} | ConvertTo-Json -Depth 5
