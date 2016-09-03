FROM haskell
RUN echo "test" | base64
COPY ./run.sh /run.sh
ENTRYPOINT ["bash", "/run.sh"]
