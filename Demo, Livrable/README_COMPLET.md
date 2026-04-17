# Borders.io - README Complet

## Présentation

**Borders.io** est un jeu multijoueur en temps réel où les joueurs contrôlent des personnages différents pour capturer du territoire. Avec des capacités uniques et des événements dynamiques, c'est un mélange d'action, de stratégie et de compétition.

## 🎮 Caractéristiques principales

- **3 personnages jouables** avec capacités distinctes :
  - 🧟 **Zombie** : expansion de terrain (+5%), ralenti (SlowCapa)
  - 🔨 **Ouvrier** : destruction de murs, multiplicateur boost bas (1.5x)
  - 🏃 **Chasseur** : ralentissement des adversaires, boost puissant (2x)

- **Gameplay en temps réel** : WebSocket, synchronisation immédiate entre joueurs
- **Événement dynamique** : **Trigger Bomb** — nettoyage orbital aléatoire à certains % de capture
- **Contrôle tactile** : d-pad directionnel, boutons boost et capacité, interface adaptée mobile
- **Visuels dynamiques** : animations boost (burst + aura), effets de nettoyage orbital

## 📋 Contenu du livrable

```
borders_io/
├── Jeu/                          # Code source complet + exécutable
│   ├── borders_io.exe           # Exécutable compilé (Windows)
│   ├── borders_io                # Exécutable compilé (Linux/Mac)
│   ├── go.mod / go.sum          # Dépendances Go
│   ├── main.go
│   ├── public/                  # Assets front (HTML, CSS, JS, images)
│   ├── network/                 # Logique serveur (WebSocket, hub)
│   ├── models/                  # Structures de données
│   ├── engine/                  # Moteur de jeu (grille, obstacles, collisions)
│   ├── player/                  # Logique joueur (mouvement, traînée)
│   ├── characters/              # Capacités des 3 personnages
│   └── events/                  # Événements (Trigger Bomb)
│
├── INDEX.md                      # Vue d'ensemble (ce fichier)
├── INSTALLATION.md              # Guide d'installation et lancement
├── MAINTENANCE.md               # Guide de maintenance et modification
└── README_COMPLET.md            # Documentation détaillée
```

## 🚀 Démarrage rapide

### Installation
```bash
# Extraire l'archive
unzip borders_io.zip
cd borders_io/Jeu

# Télécharger les dépendances (optionnel, si relance depuis source)
go mod download

# Compiler (si exécutable manquant)
go build -o borders_io.exe
```

### Lancement
```bash
# Depuis l'exécutable
./borders_io.exe

# Ou depuis le code source
go run .
```

**Serveur démarrage** : `http://localhost:8080`

### Accès au jeu
1. **Écran principal** : ouvrir `http://localhost:8080` dans un navigateur (affichage partagé)
2. **Contrôleurs mobiles** : `http://localhost:8080/controller` sur chaque téléphone/client
   - Tous les contrôleurs doivent être sur le **même réseau WiFi** que le serveur

## 🎯 Comment jouer

1. **Sélectionner un personnage** au lancement (Zombie, Ouvrier ou Chasseur)
2. **Diriger** avec les flèches ou le d-pad
3. **Capturer du territoire** en formant des boucles fermées (traînée rouge)
4. **Booster** (bouton BOOST) pour 3-5 secondes selon le personnage
5. **Utiliser sa capacité** (bouton personnalisé : ÉTENDRE, DÉTRUIRE, RALENTIR)
6. **Attendre les événements** : à 5% et tous les 2% de capture globale, une onde de choc orbital nettoie la map
7. **Gagner** : être le joueur avec le plus de territoire à la fin

## 🛠️ Technologies

- **Backend** : Go 1.20+ (réseau WebSocket, moteur jeu)
- **Frontend** : HTML5, CSS3, Vanilla JavaScript (pas de framework)
- **Communication** : WebSocket en temps réel (< 50ms latence)
- **Plateforme** : Cross-platform (Windows, Linux, Mac) + mobile (iOS, Android)

## 📖 Documentation

- **Installation & Lancement** : voir `INSTALLATION.md`
- **Maintenance & Modifications** : voir `MAINTENANCE.md`
- **Architecture complète** : voir `MAINTENANCE.md` section "Architecture générale"

## ⚠️ Notes importantes

- **Navigateur requis** : Chrome 60+, Firefox 55+, Safari 11+, Edge 79+ (WebSocket support)
- **Réseau** : un serveur Go, plusieurs clients (téléphones + affichage)
- **Performance** : testée jusqu'à 4 joueurs simultanés sans lag notable
- **Données** : aucune persistance (reset à chaque partie)

## 🐛 Support / Troubleshooting

Si le serveur ne démarre pas :
- Vérifier que le port 8080 est libre : `netstat -an | grep 8080`
- Relancer : `go run .` avec détails

Si pas de connexion depuis téléphone :
- Vérifier que serveur et clients sont sur le **même WiFi**
- Tester : `ping <server-ip>` depuis le téléphone
- Vérifier le pare-feu (port 8080 ouvert)

Voir `MAINTENANCE.md` pour détails additionnels.

## 📅 Informations projet

- **Matière** : Projet S4 — Jeu multijoueur
- **Équipe** : [Voir cahier des charges]
- **Deadline** : Mercredi 15 avril 2026, 23:59
- **Rendu** : code source + exécutable + documentation

---

**Bon jeu ! 🚀**
