{{- define "api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "api.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "api.labels" -}}
helm.sh/chart: {{ include "api.chart" . }}
{{ include "api.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "api.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "api.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "api.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "api.migrationImage" -}}
{{- printf "%s:%s" .Values.migration.image.repository (.Values.migration.image.tag | default .Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "api.terminationGracePeriodSeconds" -}}
{{- $required := add .Values.app.preShutdownDelaySeconds .Values.app.shutdownTimeoutSeconds .Values.app.gracePeriodBufferSeconds -}}
{{- if .Values.terminationGracePeriodSeconds -}}
{{- if lt (int .Values.terminationGracePeriodSeconds) (int $required) -}}
{{- fail (printf "terminationGracePeriodSeconds is %v, but must be at least %v (preShutdownDelay %vs + shutdownTimeout %vs + buffer %vs)" .Values.terminationGracePeriodSeconds $required .Values.app.preShutdownDelaySeconds .Values.app.shutdownTimeoutSeconds .Values.app.gracePeriodBufferSeconds) -}}
{{- end -}}
{{- .Values.terminationGracePeriodSeconds -}}
{{- else -}}
{{- $required -}}
{{- end -}}
{{- end -}}

{{- define "api.databaseEnv" -}}
- name: DB_URL
  valueFrom:
    secretKeyRef:
      name: {{ .Values.database.existingSecret }}
      key: {{ .Values.database.urlKey }}
{{- end -}}
