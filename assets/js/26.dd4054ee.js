(window.webpackJsonp=window.webpackJsonp||[]).push([[26],{345:function(a,s,t){"use strict";t.r(s);var e=t(25),r=Object(e.a)({},(function(){var a=this,s=a._self._c;return s("ContentSlotsDistributor",{attrs:{"slot-key":a.$parent.slotKey}},[s("h1",{attrs:{id:"quick-start"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#quick-start"}},[a._v("#")]),a._v(" Quick-start")]),a._v(" "),s("h2",{attrs:{id:"installation"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#installation"}},[a._v("#")]),a._v(" Installation")]),a._v(" "),s("h3",{attrs:{id:"macos"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#macos"}},[a._v("#")]),a._v(" MacOS")]),a._v(" "),s("p",[a._v("The preferred method is installation via "),s("a",{attrs:{href:"https://brew.sh/",target:"_blank",rel:"noopener noreferrer"}},[a._v("Homebrew"),s("OutboundLink")],1),a._v(".")]),a._v(" "),s("div",{staticClass:"language-sh extra-class"},[s("pre",{pre:!0,attrs:{class:"language-sh"}},[s("code",[a._v("brew "),s("span",{pre:!0,attrs:{class:"token function"}},[a._v("install")]),a._v(" salsadigitalauorg/shipshape/shipshape\n")])])]),s("h3",{attrs:{id:"linux"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#linux"}},[a._v("#")]),a._v(" Linux")]),a._v(" "),s("div",{staticClass:"language-sh extra-class"},[s("pre",{pre:!0,attrs:{class:"language-sh"}},[s("code",[s("span",{pre:!0,attrs:{class:"token function"}},[a._v("curl")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("-L")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("-o")]),a._v(" shipshape https://github.com/salsadigitalauorg/shipshape/releases/latest/download/shipshape-"),s("span",{pre:!0,attrs:{class:"token variable"}},[s("span",{pre:!0,attrs:{class:"token variable"}},[a._v("$(")]),s("span",{pre:!0,attrs:{class:"token function"}},[a._v("uname")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("-s")]),s("span",{pre:!0,attrs:{class:"token variable"}},[a._v(")")])]),a._v("-"),s("span",{pre:!0,attrs:{class:"token variable"}},[s("span",{pre:!0,attrs:{class:"token variable"}},[a._v("$(")]),s("span",{pre:!0,attrs:{class:"token function"}},[a._v("uname")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("-m")]),s("span",{pre:!0,attrs:{class:"token variable"}},[a._v(")")])]),a._v("\n"),s("span",{pre:!0,attrs:{class:"token function"}},[a._v("chmod")]),a._v(" +x shipshape\n"),s("span",{pre:!0,attrs:{class:"token function"}},[a._v("mv")]),a._v(" shipshape /usr/local/bin/shipshape\n")])])]),s("h3",{attrs:{id:"docker"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#docker"}},[a._v("#")]),a._v(" Docker")]),a._v(" "),s("p",[a._v("Run directly from a docker image:")]),a._v(" "),s("div",{staticClass:"language-sh extra-class"},[s("pre",{pre:!0,attrs:{class:"language-sh"}},[s("code",[s("span",{pre:!0,attrs:{class:"token function"}},[a._v("docker")]),a._v(" run "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("--rm")]),a._v(" ghcr.io/salsadigitalauorg/shipshape:latest shipshape "),s("span",{pre:!0,attrs:{class:"token parameter variable"}},[a._v("--version")]),a._v("\n")])])]),s("p",[a._v("Or add to your docker image:")]),a._v(" "),s("div",{staticClass:"language-Dockerfile extra-class"},[s("pre",{pre:!0,attrs:{class:"language-dockerfile"}},[s("code",[s("span",{pre:!0,attrs:{class:"token instruction"}},[s("span",{pre:!0,attrs:{class:"token keyword"}},[a._v("COPY")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token options"}},[s("span",{pre:!0,attrs:{class:"token property"}},[a._v("--from")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("=")]),s("span",{pre:!0,attrs:{class:"token string"}},[a._v("ghcr.io/salsadigitalauorg/shipshape:latest")])]),a._v(" /usr/local/bin/shipshape /usr/local/bin/shipshape")]),a._v("\n")])])]),s("h2",{attrs:{id:"usage"}},[s("a",{staticClass:"header-anchor",attrs:{href:"#usage"}},[a._v("#")]),a._v(" Usage")]),a._v(" "),s("p",[a._v("Create a config file. Can be as simple as:")]),a._v(" "),s("div",{staticClass:"language-yaml extra-class"},[s("pre",{pre:!0,attrs:{class:"language-yaml"}},[s("code",[s("span",{pre:!0,attrs:{class:"token comment"}},[a._v("# shipshape.yml")]),a._v("\n"),s("span",{pre:!0,attrs:{class:"token key atrule"}},[a._v("checks")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v(":")]),a._v("\n  "),s("span",{pre:!0,attrs:{class:"token key atrule"}},[a._v("file")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v(":")]),a._v("\n    "),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v("-")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token key atrule"}},[a._v("name")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v(":")]),a._v(" Illegal files\n      "),s("span",{pre:!0,attrs:{class:"token key atrule"}},[a._v("path")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v(":")]),a._v(" web\n      "),s("span",{pre:!0,attrs:{class:"token key atrule"}},[a._v("disallowed-pattern")]),s("span",{pre:!0,attrs:{class:"token punctuation"}},[a._v(":")]),a._v(" "),s("span",{pre:!0,attrs:{class:"token string"}},[a._v("'^(adminer|phpmyadmin|bigdump)?\\.php$'")]),a._v("\n")])])]),s("p",[a._v("See the "),s("a",{attrs:{href:"/config"}},[a._v("configuration")]),a._v(" documentation for more information.")]),a._v(" "),s("div",{staticClass:"language- extra-class"},[s("pre",{pre:!0,attrs:{class:"language-text"}},[s("code",[a._v('$ shipshape -h\nShipshape\n\nRun checks quickly on your project.\n\nUsage:\n  shipshape [dir]\n\nFlags:\n  -e, --error-code      Exit with error code if a failure is detected (env: SHIPSHAPE_ERROR_ON_FAILURE)\n  -d, --exclude-db      Exclude checks requiring a database; overrides any db checks specified by \'--types\'\n  -f, --file string     Path to the file containing the checks (default "shipshape.yml")\n  -h, --help            Displays usage information\n  -o, --output string   Output format [json|junit|simple|table] (env: SHIPSHAPE_OUTPUT_FORMAT) (default "simple")\n  -t, --types strings   Comma-separated list of checks to run; default is empty, which will run all checks\n  -v, --version         Displays the application version\n')])])])])}),[],!1,null,null,null);s.default=r.exports}}]);