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
Require a Secret or ConfigMap and (optionally) a list of keys inside it.
- If .keys is omitted or empty, only the resource's existence is validated.

resourceType is limited to "Secret" or "ConfigMap" when keys need to be validated, because of the "data" key used for
validation.

Usage:
  {{ include "ecosystem-core.requireSecretOrConfigMap" (dict
  		"namespace" .Release.Namespace
  		"resourceType" "Secret"
        "name" "my-secret"
        "keys" (list "username" "password")
     ) }}

  # Only check that the Secret exists:
  {{ include "ecosystem-core.requireSecretOrConfigMap" (dict "namespace" .Release.Namespace "resourceType" "Secret" "name" "my-secret") }}
*/}}
{{- define "ecosystem-core.requireSecretOrConfigMap" -}}
  {{- $ns   := required "missing namespace" .namespace -}}
  {{- $type := required "missing resource type" .resourceType -}}
  {{- $name := required "requireSecret: missing 'name' parameter" .name -}}
  {{- $keys := .keys | default (list) -}}

  {{- $obj := lookup "v1" $type $ns $name -}}
  {{- if not $obj -}}
    {{- fail (printf "%s '%s' does not exist in namespace '%s'." $type $name $ns) -}}
  {{- end -}}

  {{- if gt (len $keys) 0 -}}
    {{- $data   := (index $obj "data") | default (dict) -}}
    {{- $missing := list -}}
    {{- range $i, $k := $keys -}}
      {{- if not (hasKey $data $k) -}}
        {{- $missing = append $missing $k -}}
      {{- end -}}
    {{- end -}}
    {{- if gt (len $missing) 0 -}}
      {{- fail (printf "%s '%s' in namespace '%s' is missing required key(s): %s."
                      $type $name $ns (join ", " $missing)) -}}
    {{- end -}}
  {{- end -}}
{{- end -}}


{{- define "printCloudoguLogo" }}
{{- printf "\n" }}
...
                    ./////,
                ./////==//////*
               ////.  ___   ////.
        ,**,. ////  ,////A,  */// ,**,.
   ,/////////////*  */////*  *////////////A
  ////'        \VA.   '|'   .///'       '///*
 *///  .*///*,         |         .*//*,   ///*
 (///  (//////)**--_./////_----*//////)   ///)
  V///   '°°°°      (/////)      °°°°'   ////
   V/////(////////\. '°°°' ./////////(///(/'
      'V/(/////////////////////////////V'
{{- printf "\n" }}
{{- end }}

{{/* Renders a single Component CR from a map entry (name + component spec) */}}
{{- define "ecosystem-core.renderComponent" -}}
{{- $name := .name -}}
{{- $c := .component -}}
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: {{ $c.name | default $name }}
spec:
  name: {{ $c.name | default $name }}
  namespace: {{ $c.helmNamespace | default "k8s" }}
  version: {{ (ternary "" $c.version (eq $c.version "latest")) | quote }}
  {{- if $c.deployNamespace }}
  deployNamespace: {{ $c.deployNamespace }}
  {{- end }}
  {{- if $c.mainLogLevel }}
  mappedValues:
    mainLogLevel: {{ $c.mainLogLevel }}
  {{- end }}
  {{- with $c.valuesYamlOverwrite }}
  valuesYamlOverwrite: |-
{{ . | nindent 4 }}
  {{- end }}
{{- end }}

{{/* Renders all Component CRs from a map[string]component */}}
{{- define "ecosystem-core.renderComponentsMap" -}}
{{- $m := .map -}}
{{- range $n, $comp := $m }}
{{ include "ecosystem-core.renderComponent" (dict "name" $n "component" $comp) }}
---
{{- end }}
{{- end }}
