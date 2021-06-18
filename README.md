# Run applications

## build with docker 
`$ docker build -t my-image-hello-world .`
`$ docker container create --name jenkins-tag-cli -e PORT=1323 -e INSTANCE_ID="jenkins tag CLI" -p 1323:1323 jenkins-tag-cli`

## start container
`$ docker container start jenkins-tag-cli`


## stop container
`$ docker container stop jenkins-tag-cli`
