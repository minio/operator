
FROM openjdk:11

RUN mkdir /app
WORKDIR /app

RUN mvn pacakge

ENTRYPOINT ["java", "-jar", "./app/target/operator-sts-0.1.0.jar"]
