{{/*
Expand the name of the chart.
*/}}
{{- define "ecosystem-core.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ecosystem-core.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ecosystem-core.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ecosystem-core.labels" -}}
helm.sh/chart: {{ include "ecosystem-core.chart" . }}
{{ include "ecosystem-core.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ecosystem-core.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ecosystem-core.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ecosystem-core.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ecosystem-core.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{/*
Require a Secret and (optionally) a list of keys inside it.
- If .keys is omitted or empty, only the Secret's existence is validated.

Usage:
  {{ include "ecosystem-core.requireSecret" (dict
  		"namespace" .Release.Namespace
        "name" "my-secret"
        "keys" (list "username" "password")
     ) }}

  # Only check that the Secret exists:
  {{ include "ecosystem-core.requireSecret" (dict "namespace" .Release.Namespace "name" "my-secret") }}
*/}}
{{- define "ecosystem-core.requireSecret" -}}
  {{- $ns   := required "missing namespace" .namespace -}}
  {{- $name := required "requireSecret: missing 'name' parameter" .name -}}
  {{- $keys := .keys | default (list) -}}
  {{- $skip := default false .skip -}}

  {{- if not $skip }}
    {{- $secret := lookup "v1" "Secret" $ns $name -}}
    {{- if not $secret -}}
      {{- fail (printf "Secret '%s' does not exist in namespace '%s'." $name $ns) -}}
    {{- end -}}

    {{- if gt (len $keys) 0 -}}
      {{- $missing := list -}}
      {{- range $i, $k := $keys -}}
        {{- if not (hasKey $secret.data $k) -}}
          {{- $missing = append $missing $k -}}
        {{- end -}}
      {{- end -}}
      {{- if gt (len $missing) 0 -}}
        {{- fail (printf "Secret '%s' in namespace '%s' is missing required key(s): %s."
                        $name $ns (join ", " $missing)) -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}


