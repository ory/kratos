all:
	@npm install -d
	@cp scripts/githooks/* .git/hooks/
	@chmod -R +x .git/hooks/

colors:
	@node tests/colors.test.js


files := $(shell find . -name '*.js' ! -path "*node_modules/*")
lint:
	@node_modules/.bin/jshint ${files}

theme = $(HOME)/.spm/themes/one
documentation:
	@cp README.md _docs/index.md
	@nico build --theme=${theme}
	@cp screen-shot.png _site/

publish: clean documentation
	@ghp-import _site -p

clean:
	@rm -fr _site

server:
	@cp README.md _docs/index.md
	@nico server --theme=${theme}

.PHONY: all build test lint coverage
