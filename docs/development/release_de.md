## Release durchführen

Ein Release wird über den interaktiven Git-Flow-Release-Workflow gestartet:

```bash
make ecosystem-core-release
```

Vor dem Release sollten die Komponenten-Versionen in der `values.yaml` aktualisiert werden
(siehe [Komponenten aktualisieren](update-versions_de.md)).

### Hinweise zum Release-PR

- Die eingetragenen Komponenten-Versionen müssen zueinander passen und aufeinander abgestimmt sein.
- Hängt eine neue Version einer Komponente von einer anderen Komponente ab, die noch nicht releast ist,
  kann zunächst ein Draft-PR erstellt werden. Sobald die abhängige Komponente releast ist, wird deren
  neue Version eingetragen und der PR erst dann ins Review gegeben.

### Review und Test des Release-PRs

- Die Release-Notes der aktualisierten Komponenten ansehen und bei den Tests berücksichtigen.
- Mindestens einmal mit der Default-Konfiguration testen, ob damit alles läuft.
- Abhängig von den Release-Notes der aktualisierten Komponenten weitere Tests durchführen
  (z.B. LOP-IdP aktivieren oder andere Konfigurationen setzen).
