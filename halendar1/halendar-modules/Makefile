# Paramètres (ex : make mails N=10)
#   N       nombre de mails
#   JOURS   jours d'agenda
#   FICHIER fichier JSON (envoyer, ajouter)
N       ?= 5
JOURS   ?= 7
FICHIER ?=

BIN := bin/halendar

.DEFAULT_GOAL := help
.PHONY: help build mails envoyer agenda ajouter clean

help: ## Affiche cette aide
	@printf "Usage : make <cible> [PARAM=valeur]\n\n"
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
	@printf "\nExemples :\n"
	@echo "  make mails N=10"
	@echo "  make envoyer FICHIER=exemples/envoi.json"
	@echo "  make ajouter FICHIER=exemples/evenement.json"

build: ## Compile le programme
	@go build -o $(BIN) .

mails: build ## Derniers mails            (N=5)
	@$(BIN) mails $(N)

envoyer: build ## Envoie un mail             (FICHIER=exemples/envoi.json)
	@$(BIN) envoyer $(or $(FICHIER),exemples/envoi.json)

agenda: build ## Emploi du temps            (JOURS=7)
	@$(BIN) agenda $(JOURS)

ajouter: build ## Ajoute des RDV             (FICHIER=exemples/evenement.json)
	@$(BIN) ajouter $(or $(FICHIER),exemples/evenement.json)

clean: ## Supprime le binaire
	rm -rf bin