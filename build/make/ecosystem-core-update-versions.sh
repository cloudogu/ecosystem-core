#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

CURL_BIN="$(command -v curl || true)"

if [ -z "$CURL_BIN" ]; then
  echo "ERROR: curl ist nicht installiert"
  exit 1
fi

YAML_FILE="k8s/helm/values.yaml"
YQ_BIN=".bin/yq"
MAPPING_FILE="build/make/repo-mapping.txt"

AUTH_HEADER=""

if [ -n "${GITHUB_TOKEN:-}" ]; then
  AUTH_HEADER="Authorization: Bearer $GITHUB_TOKEN"
fi

if [ ! -x "$YQ_BIN" ]; then
  echo "ERROR: $YQ_BIN existiert nicht oder ist nicht ausführbar"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq ist nicht installiert"
  exit 1
fi

if [ ! -f "$YAML_FILE" ]; then
  echo "ERROR: Datei '$YAML_FILE' existiert nicht"
  exit 1
fi

COMPONENT_PATHS=$(
  $YQ_BIN eval '
    ..
    | select(type == "!!map" and has("version"))
    | path
    | join(".")
  ' "$YAML_FILE"
)

echo "Suche Komponenten in $YAML_FILE ..."

for PATH in $COMPONENT_PATHS; do
  COMPONENT="${PATH##*.}"

  #
  # Default: Repo == Komponentenname
  #
  REPO_NAME="$COMPONENT"

  #
  # Optionales Mapping laden
  #
  if [ -f "$MAPPING_FILE" ]; then
    while IFS='=' read -r MAP_COMPONENT MAP_REPO || [ -n "${MAP_COMPONENT:-}" ]; do
      #
      # Leere Zeilen ignorieren
      #
      if [ -z "${MAP_COMPONENT:-}" ]; then
        continue
      fi

      #
      # Kommentare ignorieren
      #
      case "$MAP_COMPONENT" in
        \#*)
          continue
          ;;
      esac

      if [ "$MAP_COMPONENT" = "$COMPONENT" ]; then
        REPO_NAME="$MAP_REPO"
        break
      fi
    done < "$MAPPING_FILE"
  fi

  echo "Hole neueste Version für $COMPONENT (Repo: $REPO_NAME) ..."

  RESPONSE=$($CURL_BIN -s \
    -H "Accept: application/vnd.github+json" \
    -H "User-Agent: update-versions-script" \
    -H "$AUTH_HEADER" \
    "https://api.github.com/repos/cloudogu/${REPO_NAME}/releases/latest" \
    || true)

  VERSION=$(printf '%s' "$RESPONSE" | $YQ_BIN -r '.tag_name // ""' 2>/dev/null || true)
  VERSION="${VERSION#v}"

  if [ -z "$VERSION" ]; then
    echo "WARNUNG: Keine Release-Version für $COMPONENT gefunden"
    continue
  fi

  echo " -> $VERSION"

  $YQ_BIN eval -i \
    ".${PATH}.version = \"${VERSION}\"" \
    "$YAML_FILE"
done

echo "Hole neueste Version für k8s-component-operator ..."

RESPONSE=$($CURL_BIN -s \
  -H "Accept: application/vnd.github+json" \
  -H "User-Agent: update-versions-script" \
  -H "$AUTH_HEADER" \
  "https://api.github.com/repos/cloudogu/k8s-component-operator/releases/latest" \
  || true)

VERSION=$(printf '%s' "$RESPONSE" | $YQ_BIN -r '.tag_name // ""' 2>/dev/null || true)

VERSION="${VERSION#v}"

if [ -n "$VERSION" ]; then
  echo " -> $VERSION"

  $YQ_BIN eval -i \
    '.k8s-component-operator.manager.image.tag = "'"$VERSION"'"' \
    "$YAML_FILE"
else
  echo "WARNUNG: Keine Version für k8s-component-operator gefunden"
fi

echo "Fertig."