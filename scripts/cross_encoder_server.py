#!/usr/bin/env python3
"""
Cross-Encoder Reranking Server for LangGraphGo

This script provides a simple HTTP server for cross-encoder based document reranking.
It can be used with the CrossEncoderReranker in Go.

Setup:
    pip install sentence-transformers flask flask-cors

Usage:
    python cross_encoder_server.py --model cross-encoder/ms-marco-MiniLM-L-6-v2 --port 8000

Then configure the Go CrossEncoderReranker with:
    config := retriever.CrossEncoderRerankerConfig{
        APIBase: "http://localhost:8000/rerank",
        ModelName: "cross-encoder/ms-marco-MiniLM-L-6-v2",
    }
"""

from flask import Flask, request, jsonify
from flask_cors import CORS
from sentence_transformers import CrossEncoder
import argparse
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)
CORS(app)

# Global model variable
model = None
model_name = None

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "ok",
        "model": model_name
    })

@app.route('/rerank', methods=['POST'])
def rerank():
    """
    Rerank documents based on query relevance.

    Expected JSON body:
    {
        "query": "search query",
        "documents": ["document 1", "document 2", ...],
        "top_n": 5 (optional),
        "model": "model-name" (optional)
    }

    Returns:
    {
        "scores": [0.95, 0.87, ...],
        "indices": [0, 2, ...]
    }
    """
    try:
        data = request.get_json()

        if not data:
            return jsonify({"error": "No JSON data provided"}), 400

        query = data.get('query')
        documents = data.get('documents', [])
        top_n = data.get('top_n', len(documents))

        if not query:
            return jsonify({"error": "Missing 'query' field"}), 400

        if not documents:
            return jsonify({"error": "Missing or empty 'documents' field"}), 400

        logger.info(f"Reranking {len(documents)} documents for query: {query[:50]}...")

        # Prepare query-document pairs
        pairs = [[query, doc] for doc in documents]

        # Get scores from cross-encoder model
        scores = model.predict(pairs).tolist()

        # Get indices sorted by score (descending)
        indexed_scores = list(enumerate(scores))
        indexed_scores.sort(key=lambda x: x[1], reverse=True)

        # Extract top N results
        top_indices = [idx for idx, _ in indexed_scores[:top_n]]
        top_scores = [score for _, score in indexed_scores[:top_n]]

        logger.info(f"Top score: {max(scores):.4f}, Average score: {sum(scores)/len(scores):.4f}")

        return jsonify({
            "scores": top_scores,
            "indices": top_indices
        })

    except Exception as e:
        logger.error(f"Error during reranking: {str(e)}")
        return jsonify({"error": str(e)}), 500

@app.route('/models', methods=['GET'])
def list_models():
    """List available pre-trained cross-encoder models"""
    models = [
        {
            "name": "cross-encoder/ms-marco-MiniLM-L-6-v2",
            "description": "Fast and accurate model for English reranking",
            "languages": ["en"],
            "size": "~80MB"
        },
        {
            "name": "cross-encoder/ms-marco-MiniLM-L-12-v2",
            "description": "Larger model with better accuracy",
            "languages": ["en"],
            "size": "~300MB"
        },
        {
            "name": "cross-encoder/mmarco-mMiniLMv2-L12-H384-v1",
            "description": "Multilingual model for various languages",
            "languages": ["en", "zh", "es", "fr", "de", "it", "pt", "ru"],
            "size": "~400MB"
        },
        {
            "name": "cross-encoder/quora-distilroberta-base",
            "description": "Good for question-answer similarity",
            "languages": ["en"],
            "size": "~250MB"
        }
    ]
    return jsonify({"models": models})

def main():
    global model, model_name

    parser = argparse.ArgumentParser(description='Cross-Encoder Reranking Server')
    parser.add_argument(
        '--model',
        type=str,
        default='cross-encoder/ms-marco-MiniLM-L-6-v2',
        help='Model name or path (default: cross-encoder/ms-marco-MiniLM-L-6-v2)'
    )
    parser.add_argument(
        '--port',
        type=int,
        default=8000,
        help='Port to run the server on (default: 8000)'
    )
    parser.add_argument(
        '--host',
        type=str,
        default='0.0.0.0',
        help='Host to bind to (default: 0.0.0.0)'
    )

    args = parser.parse_args()

    logger.info(f"Loading cross-encoder model: {args.model}")
    model_name = args.model
    model = CrossEncoder(args.model)
    logger.info("Model loaded successfully")

    logger.info(f"Starting server on {args.host}:{args.port}")
    logger.info(f"Rerank endpoint: http://{args.host}:{args.port}/rerank")
    logger.info(f"Health check: http://{args.host}:{args.port}/health")

    app.run(host=args.host, port=args.port, debug=False)

if __name__ == '__main__':
    main()
