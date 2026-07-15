pipeline {
    agent {
        label 'go'
    }

    parameters {
        booleanParam(
            name: 'RUN_TEST',
            defaultValue: true,
            description: 'Jalankan unit test'
        )
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
            when {
                expression {
                    return params.RUN_TEST
                }
            }

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
                    docker rm -f intro-ci-cd-test 2>/dev/null || true

                    docker run -d \
                        --name intro-ci-cd-test \
                        -p 2000:2000 \
                        ${IMAGE_NAME}:${IMAGE_TAG}

                    sleep 3

                    docker ps --filter "name=intro-ci-cd-test"
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
            sh 'docker rm -f intro-ci-cd-test 2>/dev/null || true'
            echo "Pipeline selesai pada node: ${env.NODE_NAME}"
        }
    }
}