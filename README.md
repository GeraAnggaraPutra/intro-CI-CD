# Intro CI/CD dengan Jenkins, Go, Docker, GitHub Webhook, dan Nexus

Pipeline yang dibuat akan melakukan proses berikut:

```text
Developer melakukan push ke branch main
        ↓
GitHub mengirim webhook
        ↓
Jenkins menerima webhook
        ↓
Pipeline dijalankan pada Go Agent
        ↓
Checkout source code
        ↓
Download dan verifikasi dependency
        ↓
Format check
        ↓
Go vet
        ↓
Unit test
        ↓
Build binary
        ↓
Archive artifact
        ↓
Build Docker image
        ↓
Menjalankan container sementara
        ↓
HTTP smoke test
        ↓
Menghapus container sementara
```

---

## 1. Teknologi yang digunakan

- Go 1.25
- Jenkins
- Jenkins SSH Build Agent
- Docker dan Docker Compose
- GitHub
- GitHub Webhook
- ngrok
- Nexus Repository
- Alpine Linux

---

## 2. Struktur project aplikasi

Contoh struktur repository:

```text
intro-CI-CD/
├── Dockerfile
├── Jenkinsfile
├── README.md
├── .dockerignore
├── go.mod
├── main.go
├── main_test.go
└── static/
    └── index.html
```

---

# Bagian A — Menjalankan aplikasi secara lokal

## 3. Persyaratan

Pastikan tools berikut sudah tersedia:

```text
Git
Go
Docker Desktop
```

Periksa dengan command satu baris:

```powershell
git --version
```

```powershell
go version
```

```powershell
docker version
```

---

## 4. Clone repository

```powershell
git clone https://github.com/GeraAnggaraPutra/intro-CI-CD.git
```

Masuk ke folder repository:

```powershell
cd intro-CI-CD
```

---

## 5. Menjalankan aplikasi tanpa Docker

Download dependency:

```powershell
go mod download
```

Jalankan unit test:

```powershell
go test -v ./...
```

Jalankan aplikasi:

```powershell
go run .
```

Buka aplikasi:

```text
http://localhost:2000
```

---

## 6. Menjalankan aplikasi dengan Docker

Build image:

```powershell
docker build -t intro-ci-cd:local .
```

Jalankan container:

```powershell
docker run --rm -p 2000:2000 --name intro-ci-cd-local intro-ci-cd:local
```

Buka aplikasi:

```text
http://localhost:2000
```

Hentikan aplikasi dengan:

```text
Ctrl+C
```

---

# Bagian B — Setup Jenkins dan Nexus

## 7. Struktur folder infrastruktur

Buat folder terpisah dari repository aplikasi:

```text
C:\jenkins-nexus\
├── compose.yml
├── .env
└── go-agent\
    └── Dockerfile
```

Buat folder dengan command satu baris:

```powershell
New-Item -ItemType Directory -Force C:\jenkins-nexus\go-agent
```

Masuk ke folder:

```powershell
cd C:\jenkins-nexus
```

---

## 8. Membuat SSH key untuk Go Agent

SSH key digunakan agar Jenkins Controller dapat login ke Go Agent.

Buat SSH key:

```powershell
ssh-keygen -t ed25519 -f "$env:USERPROFILE\.ssh\jenkins_go_agent"
```

Saat diminta passphrase, tekan `Enter` dua kali agar passphrase kosong.

File yang dibuat:

```text
C:\Users\<username>\.ssh\jenkins_go_agent
C:\Users\<username>\.ssh\jenkins_go_agent.pub
```

Private key:

```text
jenkins_go_agent
```

Public key:

```text
jenkins_go_agent.pub
```

Private key disimpan di Jenkins Credentials.

Public key dimasukkan ke container Go Agent.

---

## 9. Membuat file `.env`

Jalankan dari folder `C:\jenkins-nexus`:

```powershell
"JENKINS_AGENT_SSH_PUBKEY=$(Get-Content "$env:USERPROFILE\.ssh\jenkins_go_agent.pub")" | Set-Content .env
```

Periksa isi file:

```powershell
Get-Content .env
```

Contoh:

```dotenv
JENKINS_AGENT_SSH_PUBKEY=ssh-ed25519 AAAA... user@computer
```

Jangan memasukkan private key ke file `.env`.

---

## 10. Dockerfile Go Agent

Buat file:

```text
C:\jenkins-nexus\go-agent\Dockerfile
```

Isi:

```dockerfile
FROM jenkins/ssh-agent:jdk21

ARG GO_VERSION=1.25.4
ARG TARGETARCH=amd64

USER root

RUN apt-get update \
    && apt-get install --yes --no-install-recommends \
        ca-certificates \
        curl \
        docker-cli \
        git \
        tar \
    && curl --fail --silent --show-error --location \
        "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" \
        --output /tmp/go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm -f /tmp/go.tar.gz \
    && ln -sf /usr/local/go/bin/go /usr/local/bin/go \
    && ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt \
    && mkdir -p \
        /home/jenkins/agent \
        /home/jenkins/go \
        /home/jenkins/.cache/go-build \
    && chown -R jenkins:jenkins \
        /home/jenkins/agent \
        /home/jenkins/go \
        /home/jenkins/.cache \
    && usermod -aG root jenkins \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/opt/java/openjdk/bin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin"
ENV GOPATH="/home/jenkins/go"
ENV GOCACHE="/home/jenkins/.cache/go-build"

RUN go version \
    && git --version \
    && java -version \
    && docker --version \
    && id jenkins \
    && getent group root
```

Agent ini menyediakan:

```text
Java
Git
Go
gofmt
curl
Docker CLI
SSH server
```

---

## 11. File `compose.yml`

Buat file:

```text
C:\jenkins-nexus\compose.yml
```

Isi:

```yaml
name: jenkins-nexus

services:
  jenkins:
    image: jenkins/jenkins:2.568.1-jdk21
    container_name: jenkins
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "50000:50000"
    environment:
      TZ: Asia/Jakarta
    volumes:
      - jenkins-data:/var/jenkins_home
    networks:
      - cicd
    stop_grace_period: 120s

  go-agent:
    build:
      context: ./go-agent
      dockerfile: Dockerfile
      args:
        GO_VERSION: "1.25.4"
        TARGETARCH: amd64
    image: jenkins-go-agent:1.25.4
    container_name: jenkins-go-agent
    restart: unless-stopped
    group_add:
      - "0"
    environment:
      JENKINS_AGENT_SSH_PUBKEY: "${JENKINS_AGENT_SSH_PUBKEY}"
      TZ: Asia/Jakarta
    volumes:
      - go-agent-workspace:/home/jenkins/agent
      - go-agent-mod-cache:/home/jenkins/go
      - go-agent-build-cache:/home/jenkins/.cache/go-build
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - cicd
    stop_grace_period: 30s

  nexus:
    image: sonatype/nexus3:3.91.0
    container_name: nexus
    restart: unless-stopped
    ports:
      - "8081:8081"
      - "8082:8082"
    environment:
      TZ: Asia/Jakarta
    volumes:
      - nexus-data:/nexus-data
    networks:
      - cicd
    stop_grace_period: 120s

networks:
  cicd:
    name: cicd-network
    driver: bridge

volumes:
  jenkins-data:
    name: jenkins-data

  go-agent-workspace:
    name: go-agent-workspace

  go-agent-mod-cache:
    name: go-agent-mod-cache

  go-agent-build-cache:
    name: go-agent-build-cache

  nexus-data:
    name: nexus-data
```

Penjelasan port:

```text
8080 = Jenkins
50000 = Jenkins inbound agent port
8081 = Nexus Web UI
8082 = disiapkan untuk Docker hosted repository Nexus
```

---

## 12. Validasi Docker Compose

```powershell
docker compose config
```

Kalau tidak ada error, lanjutkan.

---

## 13. Build Go Agent

```powershell
docker compose build --no-cache go-agent
```

---

## 14. Menjalankan Jenkins, Go Agent, dan Nexus

```powershell
docker compose up -d
```

Periksa status:

```powershell
docker compose ps
```

Container yang diharapkan:

```text
jenkins
jenkins-go-agent
nexus
```

---

# Bagian C — Setup awal Jenkins

## 15. Mendapatkan password awal Jenkins

```powershell
docker compose exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword
```

Salin password yang tampil.

Buka:

```text
http://localhost:8080
```

Masukkan password, lalu pilih:

```text
Install suggested plugins
```

Buat user admin Jenkins.

---

## 16. Plugin Jenkins yang dibutuhkan

Buka:

```text
Manage Jenkins
→ Plugins
→ Available plugins
```

Install:

```text
Git
GitHub
Pipeline
SSH Build Agents
SSH Credentials
Timestamper
```

Restart Jenkins:

```powershell
docker compose restart jenkins
```

---

## 17. Menambahkan private key ke Jenkins Credentials

Tampilkan private key:

```powershell
Get-Content "$env:USERPROFILE\.ssh\jenkins_go_agent" -Raw
```

Salin seluruh isi dari:

```text
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

Buka Jenkins:

```text
Manage Jenkins
→ Credentials
→ System
→ Global credentials
→ Add Credentials
```

Isi:

```text
Kind:
SSH Username with private key

Scope:
Global

ID:
go-agent-ssh

Description:
SSH credential for Go agent

Username:
jenkins

Private Key:
Enter directly

Passphrase:
kosong
```

Simpan credential.

---

## 18. Membuat node Go Agent

Buka:

```text
Manage Jenkins
→ Nodes
→ New Node
```

Isi:

```text
Node name:
go-agent

Type:
Permanent Agent
```

Klik `Create`.

Isi konfigurasi:

```text
Number of executors:
1

Remote root directory:
/home/jenkins/agent

Labels:
go

Usage:
Only build jobs with label expressions matching this node

Launch method:
Launch agents via SSH

Host:
go-agent

Credentials:
go-agent-ssh

Host Key Verification Strategy:
Non verifying Verification Strategy

Availability:
Keep this agent online as much as possible

Port:
22
```

Simpan.

Status node harus menjadi:

```text
Online
```

---

## 19. Environment variable node

Buka:

```text
Manage Jenkins
→ Nodes
→ go-agent
→ Configure
→ Node Properties
→ Environment variables
```

Tambahkan:

```text
PATH=/opt/java/openjdk/bin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin
```

```text
GOPATH=/home/jenkins/go
```

```text
GOCACHE=/home/jenkins/.cache/go-build
```

Simpan lalu reconnect agent.

---

## 20. Memastikan Docker bisa dipakai user Jenkins

Periksa socket:

```powershell
docker compose exec go-agent ls -ln /var/run/docker.sock
```

Periksa group:

```powershell
docker compose exec go-agent getent group root
```

Hasil yang diharapkan:

```text
root:x:0:jenkins
```

Tes sebagai user Jenkins:

```powershell
docker compose exec go-agent su -s /bin/sh -c "docker version" jenkins
```

Output harus memiliki:

```text
Client:
Server:
```

Tes container:

```powershell
docker compose exec go-agent su -s /bin/sh -c "docker run --rm hello-world" jenkins
```

Target:

```text
Hello from Docker!
```

---

## 21. Tes Go Agent melalui Jenkins

Buat job:

```text
Dashboard
→ New Item
→ test-go-agent
→ Pipeline
```

Pipeline:

```groovy
pipeline {
    agent {
        label 'go'
    }

    options {
        timestamps()
        timeout(time: 5, unit: 'MINUTES')
    }

    stages {
        stage('Check Identity') {
            steps {
                sh 'whoami'
                sh 'id'
                sh 'groups'
            }
        }

        stage('Check Tools') {
            steps {
                sh 'git --version'
                sh 'go version'
                sh 'java -version'
                sh 'curl --version'
                sh 'docker version'
            }
        }

        stage('Check Docker') {
            steps {
                sh 'docker run --rm hello-world'
            }
        }
    }
}
```

Target hasil:

```text
Running on go-agent
jenkins
go version go1.25.4 linux/amd64
Client:
Server:
Hello from Docker!
Finished: SUCCESS
```

---

# Bagian D — Setup Nexus

## 22. Membuka Nexus

Buka:

```text
http://localhost:8081
```

Nexus membutuhkan waktu beberapa menit saat pertama dijalankan.

Periksa log:

```powershell
docker compose logs -f nexus
```

Tekan `Ctrl+C` setelah Nexus siap.

---

## 23. Mendapatkan password awal Nexus

```powershell
docker compose exec nexus cat /nexus-data/admin.password
```

Login dengan:

```text
Username:
admin

Password:
hasil command di atas
```

Setelah login:

```text
Ganti password admin
Nonaktifkan anonymous access untuk setup yang lebih aman
```

Pada tahap repository ini, Nexus sudah berjalan. Konfigurasi Docker hosted repository dapat ditambahkan setelah pipeline build lokal berhasil.

---

# Bagian E — Setup job Jenkins untuk repository

## 24. Membuat job pipeline

Buka:

```text
Dashboard
→ New Item
```

Isi nama:

```text
intro-ci-cd
```

Pilih:

```text
Pipeline
```

Klik `OK`.

---

## 25. Menghubungkan job dengan GitHub

Pada bagian `Pipeline`:

```text
Definition:
Pipeline script from SCM

SCM:
Git

Repository URL:
https://github.com/GeraAnggaraPutra/intro-CI-CD.git

Credentials:
None

Branch Specifier:
*/main

Script Path:
Jenkinsfile

Lightweight checkout:
aktif
```

Simpan.

---

# Bagian F — Setup GitHub Webhook dengan ngrok

## 26. Install ngrok

Periksa apakah ngrok tersedia:

```powershell
ngrok version
```

Kalau belum ada, install menggunakan Windows Package Manager:

```powershell
winget install Ngrok.Ngrok
```

Tutup dan buka kembali PowerShell.

Periksa kembali:

```powershell
ngrok version
```

---

## 27. Menambahkan authtoken ngrok

Login ke dashboard ngrok lalu salin authtoken.

Jalankan:

```powershell
ngrok config add-authtoken TOKEN_NGROK_KAMU
```

Periksa konfigurasi:

```powershell
ngrok config check
```

Jangan membagikan authtoken.

---

## 28. Menjalankan tunnel ngrok

Jenkins berjalan pada port `8080`.

Jalankan:

```powershell
ngrok http 8080
```

Contoh output:

```text
Forwarding https://contoh.ngrok-free.app -> http://localhost:8080
```

Biarkan terminal ini tetap terbuka.

Kalau terminal ditutup, tunnel berhenti.

Dashboard request ngrok tersedia di:

```text
http://localhost:4040
```

---

## 29. Mengaktifkan trigger GitHub di Jenkins

Buka job:

```text
intro-ci-cd
→ Configure
→ Triggers
```

Aktifkan:

```text
GitHub hook trigger for GITScm polling
```

Jangan aktifkan `Poll SCM` bersamaan untuk setup sederhana ini.

Simpan.

---

## 30. Menambahkan webhook di GitHub

Buka repository GitHub:

```text
Settings
→ Webhooks
→ Add webhook
```

Isi:

```text
Payload URL:
https://URL-NGROK-KAMU/github-webhook/

Content type:
application/json

Secret:
kosong untuk lab lokal

Events:
Just the push event

Active:
dicentang
```

Klik:

```text
Add webhook
```

Periksa `Recent Deliveries`.

Status sukses biasanya:

```text
200
```

---

## 31. Tes webhook

Buat empty commit:

```powershell
git commit --allow-empty -m "test: trigger Jenkins webhook" && git push origin main
```

Jenkins harus menjalankan pipeline otomatis.

Pada Console Output akan terlihat:

```text
Started by GitHub push
```

---

# Bagian G — Dockerfile aplikasi

## 32. Dockerfile aplikasi

Dockerfile repository:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main ./main
COPY --from=builder /app/static ./static

EXPOSE 2000

CMD ["./main"]
```

Penjelasan:

```text
Stage builder:
menggunakan image Go untuk compile aplikasi

Stage final:
hanya membawa binary dan folder static

Hasil:
image lebih kecil karena compiler Go tidak ikut
```

---

## 33. File `.dockerignore`

```dockerignore
.git
.gitignore
Jenkinsfile
README.md
bin
coverage.out
*.log
```

---

# Bagian H — Jenkinsfile lengkap

## 34. Pipeline final

```groovy
pipeline {
    agent {
        label 'go'
    }

    options {
        timestamps()
        timeout(time: 15, unit: 'MINUTES')
        disableConcurrentBuilds()
        skipDefaultCheckout(true)

        buildDiscarder(
            logRotator(
                numToKeepStr: '10',
                artifactNumToKeepStr: '5'
            )
        )
    }

    environment {
        CGO_ENABLED = '0'
        IMAGE_NAME = 'intro-ci-cd'
        TEST_CONTAINER_NAME = 'intro-ci-cd-test'
        DOCKER_NETWORK = 'cicd-network'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Check Environment') {
            steps {
                sh 'echo "Node: $NODE_NAME"'
                sh 'echo "Workspace: $WORKSPACE"'
                sh 'git --version'
                sh 'go version'
                sh 'docker version'
                sh 'curl --version'
            }
        }

        stage('Download Dependencies') {
            steps {
                sh 'go mod download'
            }
        }

        stage('Verify Dependencies') {
            steps {
                sh 'go mod verify'
            }
        }

        stage('Format Check') {
            steps {
                sh '''
                    unformatted="$(gofmt -l .)"

                    if [ -n "$unformatted" ]; then
                        echo "File berikut belum diformat:"
                        echo "$unformatted"
                        exit 1
                    fi

                    echo "Semua file Go sudah diformat."
                '''
            }
        }

        stage('Vet') {
            steps {
                sh 'go vet ./...'
            }
        }

        stage('Test') {
            steps {
                sh 'go test -v ./...'
            }
        }

        stage('Build Binary') {
            steps {
                sh 'mkdir -p bin && go build -o bin/intro-ci-cd .'
            }
        }

        stage('Archive Artifact') {
            steps {
                archiveArtifacts(
                    artifacts: 'bin/intro-ci-cd',
                    fingerprint: true
                )
            }
        }

        stage('Prepare Image Tag') {
            steps {
                script {
                    env.SHORT_COMMIT = sh(
                        script: 'git rev-parse --short HEAD',
                        returnStdout: true
                    ).trim()

                    env.IMAGE_TAG = "${env.BUILD_NUMBER}-${env.SHORT_COMMIT}"
                }

                echo "Docker image: ${env.IMAGE_NAME}:${env.IMAGE_TAG}"
            }
        }

        stage('Build Docker Image') {
            steps {
                sh 'docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .'
            }
        }

        stage('Inspect Docker Image') {
            steps {
                sh 'docker image inspect ${IMAGE_NAME}:${IMAGE_TAG}'
            }
        }

        stage('Smoke Test Container') {
            steps {
                sh '''
                    docker rm -f ${TEST_CONTAINER_NAME} 2>/dev/null || true

                    docker run -d \
                        --name ${TEST_CONTAINER_NAME} \
                        --network ${DOCKER_NETWORK} \
                        ${IMAGE_NAME}:${IMAGE_TAG}

                    echo "Menunggu aplikasi siap..."

                    for attempt in 1 2 3 4 5 6 7 8 9 10; do
                        if curl --fail --silent --show-error http://${TEST_CONTAINER_NAME}:2000 > /dev/null; then
                            echo "Aplikasi berhasil menerima HTTP request."
                            exit 0
                        fi

                        echo "Percobaan $attempt gagal. Mencoba kembali..."
                        sleep 2
                    done

                    echo "Aplikasi tidak siap setelah 10 percobaan."
                    docker logs ${TEST_CONTAINER_NAME}
                    exit 1
                '''
            }
        }
    }

    post {
        success {
            echo 'Pipeline berhasil.'
            echo "Docker image: ${env.IMAGE_NAME}:${env.IMAGE_TAG}"
        }

        failure {
            echo 'Pipeline gagal. Periksa stage yang merah.'
        }

        always {
            sh 'docker rm -f ${TEST_CONTAINER_NAME} 2>/dev/null || true'
            echo "Pipeline selesai pada node: ${env.NODE_NAME}"
        }
    }
}
```

---

# Bagian I — Menjalankan pipeline

## 35. Commit dan push perubahan

```powershell
git add . && git commit -m "ci: configure Jenkins pipeline" && git push origin main
```

Alur otomatis:

```text
Push ke main
→ GitHub webhook
→ Jenkins
→ Go Agent
→ pipeline dijalankan
```

---

## 36. Stage yang dijalankan

```text
Checkout
Check Environment
Download Dependencies
Verify Dependencies
Format Check
Vet
Test
Build Binary
Archive Artifact
Prepare Image Tag
Build Docker Image
Inspect Docker Image
Smoke Test Container
```

Target akhir:

```text
Finished: SUCCESS
```

---

## 37. Melihat artifact binary

Buka:

```text
Jenkins
→ intro-ci-cd
→ Build History
→ pilih build
→ Artifacts
```

Artifact:

```text
bin/intro-ci-cd
```

---

## 38. Melihat Docker image hasil Jenkins

```powershell
docker image ls intro-ci-cd
```

Contoh:

```text
REPOSITORY     TAG
intro-ci-cd    25-7b283e5
```

Tag terdiri dari:

```text
nomor build Jenkins + short commit Git
```

---

## 39. Menjalankan image hasil Jenkins

Ganti tag sesuai image yang tersedia:

```powershell
docker run --rm -p 2000:2000 --name intro-ci-cd-manual intro-ci-cd:25-7b283e5
```

Buka:

```text
http://localhost:2000
```

Hentikan dengan:

```text
Ctrl+C
```

---

# Bagian J — Command operasional

## 40. Menjalankan semua service

```powershell
docker compose up -d
```

## 41. Menghentikan semua service

```powershell
docker compose stop
```

## 42. Menjalankan kembali service

```powershell
docker compose start
```

## 43. Menghapus container tanpa menghapus volume

```powershell
docker compose down
```

## 44. Menghapus container sekaligus volume

> Peringatan: command ini menghapus data Jenkins dan Nexus.

```powershell
docker compose down -v
```

## 45. Melihat log Jenkins

```powershell
docker compose logs -f jenkins
```

## 46. Melihat log Go Agent

```powershell
docker compose logs -f go-agent
```

## 47. Melihat log Nexus

```powershell
docker compose logs -f nexus
```

## 48. Build ulang Go Agent

```powershell
docker compose build --no-cache go-agent
```

## 49. Recreate Go Agent

```powershell
docker compose up -d --force-recreate go-agent
```

---

# Bagian K — Troubleshooting

## 50. `go: not found`

Pastikan PATH node:

```text
/opt/java/openjdk/bin:/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin
```

Tes:

```powershell
docker compose exec go-agent go version
```

---

## 51. `java: not found`

Pastikan PATH memiliki:

```text
/opt/java/openjdk/bin
```

Tes:

```powershell
docker compose exec go-agent java -version
```

---

## 52. Docker socket permission denied

Periksa socket:

```powershell
docker compose exec go-agent ls -ln /var/run/docker.sock
```

Pastikan Dockerfile memiliki:

```dockerfile
usermod -aG root jenkins
```

Pastikan Compose memiliki:

```yaml
group_add:
  - "0"
```

Tes:

```powershell
docker compose exec go-agent su -s /bin/sh -c "docker version" jenkins
```

---

## 53. Smoke test tidak dapat mengakses localhost

Jangan menggunakan:

```text
http://localhost:2000
```

dari Go Agent untuk mengakses container aplikasi.

Gunakan network Docker:

```text
cicd-network
```

dan nama container:

```text
http://intro-ci-cd-test:2000
```

---

## 54. Webhook tidak menjalankan build

Periksa:

```text
GitHub
→ Repository
→ Settings
→ Webhooks
→ Recent Deliveries
```

Periksa ngrok:

```text
http://localhost:4040
```

Pastikan Payload URL:

```text
https://URL-NGROK-KAMU/github-webhook/
```

Pastikan trigger Jenkins:

```text
GitHub hook trigger for GITScm polling
```

---

## 55. URL ngrok berubah

Jalankan ulang:

```powershell
ngrok http 8080
```

Salin URL baru lalu edit webhook GitHub.

---

# Catatan keamanan

Setup ini dibuat untuk pembelajaran lokal.

Hal-hal berikut tidak direkomendasikan tanpa pengamanan tambahan di production:

```text
Host Key Verification Strategy: Non verifying
Webhook tanpa secret
Jenkins diekspos langsung melalui ngrok
Mount /var/run/docker.sock
User jenkins menjadi anggota group pemilik Docker socket
Menggunakan tag image latest tanpa versi tetap
```

Untuk production, gunakan:

```text
TLS
reverse proxy
webhook secret
credential rotation
agent terisolasi
role-based access control
image scanning
pinned image version
backup Jenkins dan Nexus
```

---

# Status pembelajaran

Bagian yang sudah dicakup:

```text
Setup Jenkins Controller
Setup Nexus
Setup Go Agent
SSH Agent connection
Go pipeline
Format check
Vet
Unit test
Build binary
Archive artifact
GitHub webhook
ngrok tunnel
Docker image build
Container smoke test
```

Tahap lanjutan yang dapat ditambahkan:

```text
Push Docker image ke Nexus
Pull image dari Nexus
Generate test report
Generate coverage report
BuildKit dan Buildx
Security scan
Staging deployment
Production deployment
Approval sebelum production
Rollback deployment
```
