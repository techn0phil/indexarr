Voici un prompt optimisé que tu peux utiliser directement :

---

**Prompt :**

> Tu es un expert en design d'interfaces web modernes. Crée une application web de gestion de médiathèque (films et séries) en HTML/CSS/JS monofichier, dans l'esprit de Sonarr et Radarr.
>
> **Architecture de navigation**
> - Sidebar fixe à gauche (210px) : logo, liens Films / Séries / Récents / Statistiques / Problèmes / Paramètres. Chaque lien Films et Séries affiche un badge avec le nombre d'éléments. Section active mise en évidence avec bordure gauche colorée (#1D9E75).
> - Topbar : barre de recherche globale pill arrondie (style Material Design, fond secondaire, focus ring vert, raccourci "/" visible à droite). Pas de titre de page, pas de bouton d'ajout.
> - En mode page détail : bouton retour + breadcrumb (ex. "Séries / Breaking Bad").
>
> **Page liste (films ou séries)**
> - Barre de filtres chips/tags cliquables (Statut, Résolution, Codec, Audio, HDR, Tri). Au clic sur un chip, une popup modale légère s'ouvre avec des options à sélection multiple (ou unique pour le tri), boutons "Effacer" et "Appliquer". Un chip actif passe en vert avec un badge compteur.
> - Toggle vue grille / vue liste en bout de barre.
> - 4 stat cards en haut du contenu (total, espace disque, % 4K, problèmes).
> - Vue grille : cards avec poster placeholder (lettre initiale), barre de statut colorée en bas (vert/orange/rouge), badges résolution/HDR/codec.
> - Vue liste : tableau dense avec colonnes Titre, Résolution, Codec, Audio, HDR, Taille, Année. Point de couleur statut en début de ligne.
>
> **Page détail film**
> - Hero : poster placeholder + titre, note TMDB avec badge source, métadonnées (année, durée, genres, statut), badges techniques (résolution, HDR, codec, audio), synopsis, boutons d'action (Rechercher upgrade, Ouvrir dans Radarr, TMDB).
> - Grille cast principal : avatar initiales + nom + rôle.
> - Tableau mediainfo multi-pistes : vidéo (codec, résolution, bitrate, HDR, fps, color space), audio par piste (codec, canaux, sample rate, bitrate, langue), sous-titres.
> - Sidebar droite : infos générales (réalisateur, studio, pays, langue, budget, box-office), fichier (taille, durée, conteneur, chemin en monospace, source, date), sources externes (TMDB ID, IMDb ID, note, popularité).
>
> **Page détail série**
> - Même hero que film, adapté : plage d'années, nombre de saisons/épisodes, statut global, bouton "Ouvrir dans Sonarr", badge source TVDB.
> - Onglets de saisons sous le hero. Au clic sur un onglet : liste des épisodes de cette saison avec numéro (E01…), titre, durée · codec · audio, badge résolution, badge "Manquant" si absent, taille, point statut. Header de saison indiquant épisodes disponibles / manquants / taille totale.
>
> **Design system**
> - Couleur principale : #1D9E75 (teal/vert). Variantes : #E1F5EE (fond actif), #5DCAA5 (bordure chip actif), #085041 (texte chip actif), #9FE1CB (badge nav actif).
> - Tous les fonds, textes et bordures utilisent des CSS variables (--color-background-primary/secondary/tertiary, --color-text-primary/secondary/tertiary, --color-border-tertiary/secondary) pour compatibilité dark mode automatique.
> - Bordures : 0.5px solid. Border-radius : var(--border-radius-md) pour éléments, var(--border-radius-lg) pour cards.
> - Police : var(--font-sans). Poids uniquement 400 et 500. Pas de gras 600/700.
> - Badges pills colorés : 4K (amber), 1080p (vert), Dolby Vision (violet), HDR10+ (bleu), Manquant (rouge).
> - Statuts : ok = #1D9E75, warning = #EF9F27, missing = #E24B4A.
> - Pas de gradient, pas d'ombre portée, pas d'emoji.
>
> **Interactivité JS**
> - Navigation entre pages (liste films, liste séries, détail film, détail série) sans rechargement.
> - Breadcrumb et bouton retour dynamiques.
> - Onglets de saisons interactifs avec rendu dynamique des épisodes.
> - Chips de filtres avec popup de sélection et mise à jour des badges compteurs.
> - Toggle grille/liste fonctionnel.

---

Ce prompt est directement réutilisable dans une nouvelle conversation — avec moi ou avec Claude Code si tu veux passer à l'implémentation réelle.

