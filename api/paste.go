package api

var LANGUAGE_NAMES_MAP = map[string]string{
	"Plaintext":                    "plaintext",
	"Python":                       "python",
	"C#":                           "csharp",
	"C":                            "c",
	"C++":                          "cpp",
	"Java":                         "java",
	"JavaScript":                   "javascript",
	"JSON":                         "json",
	"Go":                           "go",
	"SQL":                          "sql",
	"1C":                           "1c",
	"4D":                           "4d",
	"ABAP":                         "sap-abap",
	"ABNF":                         "abnf",
	"Access logs":                  "accesslog",
	"Ada":                          "ada",
	"Apex":                         "apex",
	"Arduino (C++ w/Arduino libs)": "arduino",
	"ARM assembler":                "armasm",
	"AVR assembler":                "avrasm",
	"ActionScript":                 "actionscript",
	"Alan IF":                      "alan",
	"Alan":                         "ln",
	"AngelScript":                  "angelscript",
	"Apache":                       "apache",
	"AppleScript":                  "applescript",
	"Arcade":                       "arcade",
	"AsciiDoc":                     "asciidoc",
	"AspectJ":                      "aspectj",
	"AutoHotkey":                   "autohotkey",
	"AutoIt":                       "autoit",
	"Awk":                          "awk",
	"Ballerina":                    "ballerina",
	"Bash":                         "bash",
	"Basic":                        "basic",
	"BBCode":                       "bbcode",
	"Blade (Laravel)":              "blade",
	"BNF":                          "bnf",
	"BQN":                          "bqn",
	"Brainfuck":                    "brainfuck",
	"C/AL":                         "cal",
	"C3":                           "c3",
	"Cache Object Script":          "cos",
	"Candid":                       "candid",
	"CMake":                        "cmake",
	"COBOL":                        "cobol",
	"CODEOWNERS":                   "codeowners",
	"Coq":                          "coq",
	"CSP":                          "csp",
	"CSS":                          "css",
	"Cap’n Proto":                  "capnproto",
	"Chaos":                        "chaos",
	"Chapel":                       "chapel",
	"Cisco CLI":                    "cisco",
	"Clojure":                      "clojure",
	"CoffeeScript":                 "coffeescript",
	"CpcdosC+":                     "cpc",
	"Crmsh":                        "crmsh",
	"Crystal":                      "crystal",
	"cURL":                         "curl",
	"Cypher (Neo4j)":               "cypher",
	"D":                            "d",
	"Dafny":                        "dafny",
	"Dart":                         "dart",
	"Delphi":                       "dpr",
	"Diff":                         "diff",
	"Django":                       "django",
	"DNS Zone file":                "dns",
	"Dockerfile":                   "dockerfile",
	"DOS":                          "dos",
	"dsconfig":                     "dsconfig",
	"DTS (Device Tree)":            "dts",
	"Dust":                         "dust",
	"Dylan":                        "dylan",
	"EBNF":                         "ebnf",
	"Elixir":                       "elixir",
	"Elm":                          "elm",
	"Erlang":                       "erlang",
	"Excel":                        "excel",
	"Extempore":                    "extempore",
	"F#":                           "fsharp",
	"FIX":                          "fix",
	"Flix":                         "flix",
	"Fortran":                      "fortran",
	"FunC":                         "func",
	"G-Code":                       "gcode",
	"Gams":                         "gams",
	"GAUSS":                        "gauss",
	"GDScript":                     "godot",
	"Gherkin":                      "gherkin",
	"Glimmer and EmberJS":          "hbs",
	"GN for Ninja":                 "gn",
	"Grammatical Framework":        "gf",
	"Golo":                         "golo",
	"Gradle":                       "gradle",
	"GraphQL":                      "graphql",
	"Groovy":                       "groovy",
	"GSQL":                         "gsql",
	"HTML, XML":                    "html",
	"HTTP":                         "http",
	"Haml":                         "haml",
	"Handlebars":                   "handlebars",
	"Haskell":                      "haskell",
	"Haxe":                         "haxe",
	"Hy":                           "hy",
	"Ini, TOML":                    "toml",
	"Inform7":                      "inform7",
	"IRPF90":                       "irpf90",
	"Iptables":                     "iptables",
	"JSONata":                      "jsonata",
	"Jolie":                        "jolie",
	"Julia":                        "julia",
	"Julia REPL":                   "julia-repl",
	"Kotlin":                       "kotlin",
	"LaTeX":                        "tex",
	"Leaf":                         "leaf",
	"Lean":                         "lean",
	"Lasso":                        "lasso",
	"Less":                         "less",
	"LDIF":                         "ldif",
	"Lisp":                         "lisp",
	"LiveCode Server":              "livecodeserver",
	"LiveScript":                   "livescript",
	"LookML":                       "lookml",
	"Lua":                          "lua",
	"Luau":                         "luau",
	"Macaulay2":                    "macaulay2",
	"Makefile":                     "makefile",
	"Markdown":                     "markdown",
	"Mathematica":                  "mathematica",
	"Matlab":                       "matlab",
	"Maxima":                       "maxima",
	"Maya Embedded Language":       "mel",
	"Mercury":                      "mercury",
	"MIPS Assembler":               "mips",
	"Mint":                         "mint",
	"Mirth":                        "mirth",
	"mIRC Scripting Language":      "mirc",
	"Mizar":                        "mizar",
	"MKB":                          "mkb",
	"MLIR":                         "mlir",
	"Mojolicious":                  "mojolicious",
	"Monkey":                       "monkey",
	"Moonscript":                   "moonscript",
	"Motoko":                       "motoko",
	"N1QL":                         "n1ql",
	"NSIS":                         "nsis",
	"Never":                        "never",
	"Nginx":                        "nginx",
	"Nim":                          "nim",
	"Nix":                          "nix",
	"Oak":                          "oak",
	"Object Constraint Language":   "ocl",
	"OCaml":                        "ocaml",
	"Objective C":                  "objectivec",
	"OpenGL Shading Language":      "glsl",
	"OpenSCAD":                     "openscad",
	"Oracle Rules Language":        "ruleslanguage",
	"Oxygene":                      "oxygene",
	"PF":                           "pf",
	"PHP":                          "php",
	"Papyrus":                      "papyrus",
	"Parser3":                      "parser3",
	"Perl":                         "perl",
	"Phix":                         "phix",
	"Pine Script":                  "pine",
	"Pony":                         "pony",
	"PostgreSQL & PL/pgSQL":        "pgsql",
	"PowerShell":                   "powershell",
	"Processing":                   "processing",
	"Prolog":                       "prolog",
	"Properties":                   "properties",
	"Protocol Buffers":             "proto",
	"Puppet":                       "puppet",
	"Python profiler results":      "profile",
	"Python REPL":                  "python-repl",
	"Q#":                           "qsharp",
	"Q":                            "k",
	"QML":                          "qml",
	"R":                            "r",
	"Razor CSHTML":                 "cshtml",
	"ReasonML":                     "reasonml",
	"Rebol & Red":                  "redbol",
	"RenderMan RIB":                "rib",
	"RenderMan RSL":                "rsl",
	"ReScript":                     "rescript",
	"RiScript":                     "risc",
	"RISC-V Assembly":              "riscv",
	"Roboconf":                     "graph",
	"Robot Framework":              "robot",
	"RPM spec files":               "rpm-specfile",
	"Ruby":                         "ruby",
	"Rust":                         "rust",
	"RVT Script":                   "rvt",
	"SAS":                          "SAS",
	"SCSS":                         "scss",
	"STEP Part 21":                 "p21",
	"Scala":                        "scala",
	"Scheme":                       "scheme",
	"Scilab":                       "scilab",
	"SFZ":                          "sfz",
	"Shape Expressions":            "shexc",
	"Shell":                        "shell",
	"Smali":                        "smali",
	"Smalltalk":                    "smalltalk",
	"SML":                          "sml",
	"Solidity":                     "solidity",
	"Splunk SPL":                   "spl",
	"Stan":                         "stan",
	"Stata":                        "stata",
	"Structured Text":              "iecst",
	"Stylus":                       "stylus",
	"SubUnit":                      "subunit",
	"Supercollider":                "supercollider",
	"Svelte":                       "svelte",
	"Swift":                        "swift",
	"Tcl":                          "tcl",
	"Terraform (HCL)":              "terraform",
	"Test Anything Protocol":       "tap",
	"Thrift":                       "thrift",
	"Toit":                         "toit",
	"TP":                           "tp",
	"Transact-SQL":                 "tsql",
	"TTCN-3":                       "ttcn",
	"Twig":                         "twig",
	"TypeScript":                   "typescript",
	"Unicorn Rails log":            "unicorn-rails-log",
	"Unison":                       "unison",
	"VB.Net":                       "vbnet",
	"VBA":                          "vba",
	"VBScript":                     "vbscript",
	"VHDL":                         "vhdl",
	"Vala":                         "vala",
	"Verilog":                      "verilog",
	"Vim Script":                   "vim",
	"WGSL":                         "wgsl",
	"X#":                           "xsharp",
	"X++":                          "axapta",
	"x86 Assembly":                 "x86asm",
	"x86 Assembly (AT&T)":          "x86asmatt",
	"XL":                           "xl",
	"XQuery":                       "xquery",
	"YAML":                         "yml",
	"ZenScript":                    "zenscript",
	"Zephir":                       "zephir",
	"Zig":                          "zig",
	"":                             "",
}

func isSupportedLanguage(lang string) bool {
	for _, v := range LANGUAGE_NAMES_MAP {
		if v == lang && lang != "" {
			return true
		}
	}
	return false
}
