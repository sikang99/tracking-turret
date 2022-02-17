#
# Makefile
# 
usage:
	@echo "usage: make [edit|build|run]"


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
