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

        stage('Check Tools') {
            steps {
                sh 'git --version'
                sh 'go version'
            }
        }

        stage('Build') {
            steps {
                echo 'Building application...'
            }
        }
    }
}