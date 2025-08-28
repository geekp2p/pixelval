package gateway

import _ "embed"

//go:embed ../../web/index.html
var IndexHTML []byte

//go:embed ../../web/app.js
var AppJS []byte

//go:embed ../../web/style.css
var AppCSS []byte
