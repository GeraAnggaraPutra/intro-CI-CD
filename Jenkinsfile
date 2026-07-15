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

    triggers {
        pollSCM('* * * * *')
    }

    options {
        timestamps()
        timeout(time: 10, unit: 'MINUTES')
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
                sh 'git --version'
                sh 'go version'
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
                        echo "File belum diformat:"
                        echo "$unformatted"
                        exit 1
                    fi
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

        stage('Build') {
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
    }

    post {
        success {
            echo 'CI Go berhasil.'
        }

        failure {
            echo 'CI Go gagal. Periksa stage yang gagal.'
        }

        always {
            echo "Pipeline selesai pada node: ${env.NODE_NAME}"
        }
    }
}