deploy:
	gcloud functions deploy pokebot \
	--gen2 \
	--region=us-west1 \
	--runtime=go122 \
	--source=. \
	--entry-point=Publish \
	--trigger-http --env-vars-file=env.yml