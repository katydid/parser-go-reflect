# Copyright 2013 Walter Schulze
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: nuke gofmt build test

all: nuke build test

checklicense:
	go get github.com/awalterschulze/checklicense
	checklicense . \
	doc.go \
	tools/tools.go \
	.svg \
	.txt

test:
	go test ./...

build:
	go build ./...

install:
	go install ./...

bench:
	go test -test.v -test.run=XXX -test.bench=. ./...

gofmt:
	gofmt -l -s -w .

travis:
	make all
	make checklicense
	make diff

diff:
	git diff --exit-code .
