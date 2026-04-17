# Guide de maintenance - Borders.io

## Architecture générale

```
Jeu/
├── main.go                      # Point d'entrée principal
├── go.mod / go.sum            # Dépendances Go
├── public/                     # Assets front (HTML, CSS, JS, images)
│   ├── index.html             # Page jeu (affichage principal)
│   ├── controller.html        # Page contrôleur (joueurs)
│   ├── *.css                  # Styles
│   └── img/                   # Images et vidéos
├── network/                    # Logique réseau (WebSocket)
│   ├── hub.go                 # Centre de gestion du hub
│   ├── client.go              # Gestion des clients WS
│   └── router.go              # Routes HTTP
├── models/                     # Structures de données
│   └── models.go              # Player, UpdateMessage, etc.
├── engine/                     # Moteur de jeu
│   ├── grid.go                # Grille et capture
│   ├── obstacles.go           # Murs et collisions
│   ├── combat.go              # Collisions joueur/joueur
│   └── ...
├── player/                     # Logique joueur
│   └── player.go              # Mouvement et traînée
├── characters/                 # Capacités des personnages
│   ├── capacites/ouvrier/
│   ├── capacites/chasseur/
│   └── capacites/zombie/
└── events/                     # Événements dynamiques
    └── triggerbomb/           # Event nettoyage orbital
```

## Points d'intégration clés

### 1. Ajouter une capacité personnage

**Localisation** : `characters/capacites/{character}/ability.go`

```go
// Dans ability.go, implémenter la fonction ExecteAttack
func ExecuteAttackPerso(p *models.Player) bool {
    // Logique spécifique du personnage
    return true // true si succès, false sinon
}
```

**Puis** : brancher dans `network/client.go`, fonction `readPump()`, section `use_ability`.

### 2. Ajouter un nouvel événement

**Localisation** : `events/{event}/logic.go`

Structure requise :
```go
type Manager struct { /* état */ }
type EventMessage struct { /* données à envoyer */ }

func NewManager() *Manager { /* init */ }
func (m *Manager) Evaluate(...) *EventMessage { /* logique */ }
func (m *Manager) Reset() { /* reset pour nouvelle partie */ }
```

**Intégration backend** : dans `network/hub.go`, GameLoop(), ajouter un appel à `Evaluate()`.

**Intégration front** : dans `public/index.html`, `socket.onmessage()`, traiter le type d'event.

### 3. Modifier les paramètres du jeu

| Paramètre | Localisation | Exemple |
|-----------|--------------|---------|
| **Vitesse de base** | `models/models.go`, `CharacterStats` | `BaseSpeed: 3.5` |
| **Multiplicateur boost** | `models/models.go`, `CharacterStats` | `BoostMultiplier: 2.0` |
| **Durée boost** | `models/models.go`, `CharacterStats` | `BoostDuration: 150` (ticks à 33ms) |
| **Seuil trigger bomb** | `network/hub.go`, `NewManager(5, 2)` | `5` = départ à 5% capture, `2` = paliers de 2% |
| **Cooldown capacité** | `models/models.go`, `CharacterStats` | `CapacityCooldown: 5` (secondes) |
| **Résolution grille** | `engine/grid.go`, constantes | `SCALE = 10` pixels par case |

### 4. Déboguer une partie

**Logs serveur** : le serveur affiche les événements clés dans le terminal.

**DevTools navigateur** : 
- Ouvrir F12 → Console
- Vérifier les messages WS reçus (Network → WS)
- Vérifier les erreurs JS

### 5. Modifier l'UI front

| Zone | Fichier | Objectif |
|------|---------|----------|
| **Styles globaux** | `public/general.css` | Couleurs, typos, valeurs CSS |
| **Écran de jeu** | `public/index.html` + `index.css` | Lobby, canvas, layout |
| **Contrôleur** | `public/controller.html` + `controllerGame.css` | Boutons, leaderboard, d-pad |
| **Animations** | `public/index.html` (JavaScript) | FX boost, FX events, rendering |

## Problèmes courants

### Le serveur démarre mais pas de connexion WebSocket
- Vérifier que `network/router.go` expose bien la route `/ws`.
- Vérifier que le port 8080 est accessible (`netstat -an | grep 8080`).
- Vérifier la console navigateur pour les erreurs WebSocket.

### Les joueurs ne voient pas les murs
- Vérifier que `engine/obstacles.go` peuple bien `MapObstacles`.
- Vérifier que `network/hub.go` envoie `Obstacles` dans `UpdateMessage`.
- Vérifier que `public/index.html` dessine les obstacles (ligne `cachedWall.forEach(...)`).

### Les capacités ne fonctionnent pas
- Vérifier le cooldown : `time.Now().After(p.NextCapacityTime)` dans `network/client.go`.
- Vérifier que le personnage est bien défini : `p.Character == "ouvrier"` etc.
- Vérifier que la fonction spécifique existe et retourne `true` en succès.

### L'event trigger bomb ne se déclenche pas
- Vérifier le seuil dans `network/hub.go` : `NewManager(5, 2)`.
- Vérifier que l'event manager est appelé toutes les 250ms dans GameLoop.
- Vérifier les logs serveur pour les messages `orbital_cleanup_start`.
- Front : vérifier la fonction `launchOrbitalFx()` et les animations FX.

## Tests recommandés

1. **Gameplay basique** : lancer 2-3 joueurs, vérifier que les territoires se capturent.
2. **Capacités** : tester chaque personnage, vérifier les cooldowns.
3. **Événements** : atteindre 5% de capture, observer le nettoyage orbital.
4. **Performance** : tester avec 4+ joueurs simultanés, vérifier les FPS.
5. **Mobile** : tester le contrôleur sur téléphones (iOS, Android).

## Checklist pour commit

- [ ] Code compilé sans erreurs : `go build`
- [ ] Tests simples passent : 2 joueurs jouent une partie complète
- [ ] Pas de logs d'erreur dans le terminal
- [ ] Console navigateur sans erreurs rouge
- [ ] Les assets (images, vidéos) sont présents et servies
- [ ] Documentation updated si changement d'architecture
