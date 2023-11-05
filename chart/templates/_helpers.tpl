{{/*
Expand the name of the chart.
*/}}
{{- define "k8s-secret-replicator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8s-secret-replicator.fullname" -}}
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
{{- define "k8s-secret-replicator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8s-secret-replicator.labels" -}}
helm.sh/chart: {{ include "k8s-secret-replicator.chart" . }}
{{ include "k8s-secret-replicator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8s-secret-replicator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8s-secret-replicator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8s-secret-replicator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "k8s-secret-replicator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Template image tag.
*/}}
{{- define "image-template" }}
{{- if .repository -}}
{{- printf "%s/%s/%s:%s" (required (printf "missing %s.%s" . .registry) (trimSuffix "/" .registry)) (trimAll "/" .repository) (required (printf "missing %s.%s" . .name) (trimAll "/" .name)) (required (printf "missing %s.%s" . .tag) .tag) }}
{{- else -}}
{{- printf "%s/%s:%s" (required (printf "missing %s.%s" . .registry) (trimSuffix "/" .registry)) (required (printf "missing %s.%s" . .name) (trimAll "/" .name)) (required (printf "missing %s.%s" . .tag) .tag) }}
{{- end }}
{{- end }}
