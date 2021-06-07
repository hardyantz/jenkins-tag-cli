pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        echo 'Building..'
        echo "text ${env.BRANCH}"
        echo "text ${env.PARAM2}"
      }
    }

    stage('Test') {
      steps {
        echo 'Testing..'
      }
    }

    stage('Deploy') {
      steps {
        echo 'Deploying....'
      }
    }

  }
}