# Microservice REST API en Go avec Redis

## Description
Ce projet est un microservice REST API en Go, conçu pour la gestion des commandes et l'intégration avec Redis comme base de données. Développé avec la modularité en tête, ce microservice peut facilement s'adapter à d'autres systèmes de base de données. Ce projet représente ma découverte de la créattion de microservices scalables et maintenables en Go.

## Fonctionnalités
- **Gestion des commandes :** Permet de créer, lire, mettre à jour et supprimer des commandes stockées dans Redis.
- **Architecture REST :** Microservice RESTful pour une intégration facile avec d'autres systèmes ou front-ends.
- **Modularité :** Conçu pour faciliter le remplacement ou la mise à jour des composants, tels que la base de données.
- **Logging :** Middleware intégré pour le suivi des requêtes et des réponses.

## Technologies Utilisées
- **Go** : Pour le développement du backend.
- **Chi Router** : Pour la gestion des routes et des middlewares.
- **Redis** : Comme système de stockage principal.
