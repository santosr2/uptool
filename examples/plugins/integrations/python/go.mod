module github.com/santosr2/uptool/examples/plugins/python

go 1.25

require github.com/santosr2/uptool v0.2.0-alpha20251130

require golang.org/x/text v0.25.0 // indirect

// Better local development
replace github.com/santosr2/uptool => ../../../..
