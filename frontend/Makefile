# include .env

# ============================================================================= #
# HELPERS	  																	#
# ============================================================================= #
.PHONY: help
help:
	@echo '*** Commands:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]


# ============================================================================= #
# DEPLOYMENT  																	#
# ============================================================================= #

## deploy (optional version=X.X.X): deploy the app to production
.PHONY: deploy
deploy:
	@read -p "Enter commit message: " MESSAGE; \
	flutter build web && \
	rsync -av --exclude='.git' --exclude='.last_build_id' --delete ./build/web/ ../frontend_build/ && \
	cd ../frontend_build && \
	git add . && \
	git commit -m "$$MESSAGE" && \
	git push && \
	cd ../frontend
	@if [ "x$(version)" != "x" ]; then \
		make -s -C ../backend frontend/version version=$(version); \
	fi


## deploy2: deploy the app to test repo
.PHONY: deploy2
deploy2:
	@read -p "Enter commit message: " MESSAGE; \
	flutter build web && \
	rsync -av --exclude='.git' --exclude='.last_build_id' --delete ./build/web/ ../halendar2_front/ && \
	cd ../halendar2_front && \
	git add . && \
	git commit -m "$$MESSAGE" && \
	git push && \
	cd ../halendar_front

## test: testing something
.PHONY: test
test:
	flutter run -d chrome --profile --web-port 9090


## deploy/android: deploy/android the app to production
.PHONY: deploy/android
deploy/android:
	@read -p "Enter version: " VERSION; \
	flutter build apk && \
	mv ./build/app/outputs/flutter-apk/app-release.apk "./build/app/outputs/flutter-apk/$$VERSION.apk"
