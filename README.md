# Filewise / DFAT : Directory File Analysis Tool

## Developer initial instructions:
1) Install the Go Lang Compiler
2) In VS Code Extension Store Install the GO Linter Addon
3) In the VSCode Extension store install the code runner Extension
4) Once that install inside the terminal CD into the current repo and run the command: *go get github.com/saintfish/chardet*
5) The code is ready to run

## Build Excecutable:
1) Starting from where the developer instruction ends, Open the terminal inside the repo Directory
2) Run the command: go build main.go

## Run Excecutable:
1) in the terminal inside the repo folder run "main.exe"


## User Instructions:
1) run the main.exe inside the repo directory inside a terminal or from the folder: *default behavior is to scan all files in the current directory*
2) Incase the user needs to specify an alternative directory run this command in powershell: "./main.exe -p ENTER_PATH_HERE" place full path *ENTER_PATH_HERE*

## File Output Expectation:

```json
[
  {
    "name": ".gitignore",
    "path": "C:/Users/cwedderburn/repos/FileWise/.gitignore",
    "size": 0,
    "ext": "gitignore",
    "encoding": "UTF-8",
    "is_binary": false
  },
  {
    "name": "README.md",
    "path": "C:/Users/cwedderburn/repos/FileWise/README.md",
    "size": 0.083984375,
    "ext": "md",
    "encoding": "ISO-8859-1",
    "is_binary": false
  },
  {
    "name": "FETCH_HEAD",
    "path": "C:/Users/cwedderburn/repos/FileWise/.git/FETCH_HEAD",
    "size": 0,
    "ext": "{BLANK}",
    "encoding": "UTF-8",
    "is_binary": false
  },
  {
    "name": "HEAD",
    "path": "C:/Users/cwedderburn/repos/FileWise/.git/HEAD",
    "size": 0.0224609375,
    "ext": "{BLANK}",
    "encoding": "ISO-8859-1",
    "is_binary": false
  }
]
```json

