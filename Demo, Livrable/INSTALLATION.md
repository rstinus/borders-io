# Guide d'installation - Borders.io

## Prérequis

- **Go 1.20+** ([télécharger](https://golang.org/dl/))
- **Navigateur moderne** (Chrome, Firefox, Safari, Edge)
- **Terminal/CMD** (avec accès au répertoire du projet)

## Installation depuis la source

### 1. Extraire l'archive
```bash
unzip borders_io.zip
cd borders_io/Jeu
```

### 2. Télécharger les dépendances
```bash
go mod download
# ou : go mod tidy
```

### 3. Compiler le projet
```bash
go build -o borders_io.exe
# Sur Linux/Mac : go build -o borders_io
```

**Résultat** : un exécutable `borders_io.exe` est généré.

## Lancement du serveur

### Depuis l'exécutable
```bash
./borders_io.exe
# Ou sur Windows : borders_io.exe
```

**Sortie attendue** :
```
Serveur démarré sur http://localhost:8080
```

### Depuis le code source (développement)
```bash
go run .
```

## Accès au jeu

- **Écran principal (affichage)** : ouvrir `http://localhost:8080` dans un navigateur
- **Contrôleur mobile** : `http://localhost:8080/controller` depuis chaque téléphone/client

## Troubleshooting

| Problème | Solution |
|----------|----------|
| Port 8080 déjà utilisé | Modifier le port dans `main.go` ligne `http.ListenAndServe(":8080", ...)` |
| Module `Jeu/...` introuvable | Vérifier que `go.mod` existe à la racine du projet Jeu |
| Erreur de compilation | Relancer `go mod download && go mod tidy` |
| Assets CSS/images manquants | Vérifier que le dossier `public/` existe avec ses fichiers |

## Notes

- Le serveur fonctionne sur `localhost:8080` par défaut.
- Plusieurs navigateurs peuvent se connecter simultanément (pas de limite).
- Les fichiers statiques (HTML, CSS, images) sont servis depuis le dossier `public/`.
