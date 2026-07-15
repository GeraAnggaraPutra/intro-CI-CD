pipeline {
    agent any

    options {
        timestamps()
        timeout(time: 10, unit: 'MINUTES')
        disableConcurrentBuilds()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Inspect Workspace') {
            steps {
                sh 'pwd'
                sh 'ls -la'
            }
        }

        stage('Build') {
            steps {
                echo 'Building intro-CI-CD...'
            }
        }

        stage('Test') {
            steps {
                echo 'Testing intro-CI-CD...'
            }
        }
    }

    post {
        success {
            echo 'Pipeline berhasil'
        }

        failure {
            echo 'Pipeline gagal. Periksa Console Output.'
        }

        always {
            echo 'Pipeline selesai'
        }
    }
}