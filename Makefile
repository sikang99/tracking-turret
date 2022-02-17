#
# Makefile
# 
PROG=tracking-turret
VERSION=1.0.0.1
usage:
	@echo "usage: make [edit|build|run]"

build b:
	go build -o $(PROG) main.go

edit e:
	vi main.go

run r:
	./$(PROG)
#----------------------------------------------------------------------------------
git g:
	@echo "make (git:g) [update|store]"
git-reset gr:
	git reset --soft HEAD~1
git-update gu:
	git add .
	git commit -a -m "$(VERSION),$(USER)"
	git push
git-store gs:
	git config credential.helper store
#----------------------------------------------------------------------------------
