# random-pokemon-publisher a.k.a PokeBot

This project was born from my desire to create a bot for the Bluesky social media platform and my love of Pokemon.

It leverages Bluesky's publicly available API and the publicly available PokeAPI.

This bot resides as a Google Cloud Run function on Google Cloud Platform. It has a bucket attached to it such that it once a Pokemon has been published, it creates an object in the bucket. The next time the job runs, the bucket is checked to ensure that the randomly selected Pokemon has not already been published.
