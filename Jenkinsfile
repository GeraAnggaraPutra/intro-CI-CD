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