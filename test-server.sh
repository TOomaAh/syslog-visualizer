#!/bin/bash

# Script de test pour le serveur syslog-visualizer
# Ce script envoie des messages syslog de test et vérifie qu'ils sont bien reçus via l'API

echo "🚀 Test du serveur Syslog Visualizer"
echo "====================================="
echo ""

# Attendre que le serveur démarre
sleep 2

# Envoyer quelques messages syslog via UDP
echo "📤 Envoi de messages syslog via UDP..."
echo "<34>Oct 11 22:14:15 server1 su[1234]: 'su root' failed for user" | nc -u -w1 localhost 514
echo "<13>Feb  5 17:32:18 server2 myapp: Application started successfully" | nc -u -w1 localhost 514
echo "<86>Dec  1 08:30:00 server3 kernel: Out of memory warning" | nc -u -w1 localhost 514

# Attendre un peu pour que les messages soient traités
sleep 1

# Interroger l'API pour vérifier que les messages ont été reçus
echo ""
echo "📥 Récupération des messages via l'API..."
curl -s http://localhost:8080/api/syslogs | jq '.'

echo ""
echo "✅ Test terminé !"
echo ""
echo "Pour tester manuellement :"
echo "  - Envoyer un message: echo \"<34>Oct 11 22:14:15 test su: test\" | nc -u localhost 514"
echo "  - Voir les messages: curl http://localhost:8080/api/syslogs | jq"
echo "  - Health check: curl http://localhost:8080/api/health"
