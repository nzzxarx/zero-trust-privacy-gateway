# Hybrid Zero-Trust PII Privacy Gateway 🛡️

A high-performance, two-tier hybrid data scrubbing gateway engineered to intercept, sanitize, and redact sensitive Personally Identifiable Information (PII) and custom regional identifiers (like Moroccan CNIE alphanumerics) before processing unstructured logs into public/enterprise LLM pipelines. 

---

## 🏗️ Architecture Blueprint

The system splits scanning orchestration into a decoupled pipeline to achieve ultra-low latencies (<6ms runtime latency):

1. **Deterministic Edge Layer (Go Engine):** A high-throughput concurrent Go router that catches obvious structured data parameters (emails, phone sequences, credit cards) using specialized bounded regular expressions directly at the network boundary.
2. **Semantic Context Layer (Python + IBM watsonx.ai):** When regular expression boundaries are ambiguous (e.g., messy or incomplete identification formats), the Go engine drops down asynchronously to a lightweight Python microservice invoking the **ibm-watsonx-ai SDK**. This layer leverages enterprise **Granite models** to identify and strip tokens based on linguistic context rather than rigid rules.

---

## 🛠️ Tech Stack

- **Edge Router & Filtering Engine:** Golang (Bounded channel architecture, Sync registries)
- **Contextual NLP Inference Node:** Python 3.12, Flask, IBM watsonx.ai SDK, Granite Models

---

## 🚀 Deployment Guide

### 1. Initialize the IBM watsonx Inference Service
```bash
cd ai_engine
python -m venv venv
source venv/bin/activate  # On Windows use: venv\Scripts\activate
pip install -r requirements.txt
python app.py
