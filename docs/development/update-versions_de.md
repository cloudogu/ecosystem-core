## Komponenten aktualisieren

Um die aktuellsten Komponenten Versionen in die `values.yaml` automatisiert eintragen zu können, 
kann das Maketarget `make update-ecosystem-versions` verwendet werden.

Dies prüft die cloudogu-Repos die zum Namen der Komponenten gehören und trägt die neuste Version in der Yaml-Datei ein
sofern sich die Version geändert hat.

Im Log-Output des Targets findet sich dann auch die korrekte Commit-Message mit Komponentenname, alter und neuer Version.

Einige Komponenten liegen in Repos, deren Name nicht der Komponente entspricht (z.B  CRD-Komponenten liegen in lib-Repos).
Für diese Fälle kann die `repo-mapping.txt` angepasst werden.

Um auf git per API zuzugreifen benötigt man einen GIT_TOKEN.
Diesen kann man in seiner .env Datei eintragen oder dem Maketarget mitgeben

`GIT_TOKEN=1234567890 make update-ecosystem-versions`