# Trusted LLM Model Evaluation Example

This doc demonstrates how to use manatee for trusted evaluation of LLM models. Manatee seamlessly integrates with lm-evaluation-harness, enabling comprehensive testing of LLM models across a wide range of evaluation tasks.

Scenario:
Suppose a model provider owns a proprietary LLM model. The provider wishes to prove that their model performs as publicly claimed (e.g., in terms of fairness or accuracy). This evaluation process is divided into two stages: 
- Stage 1: The script runs on a mock (fake) model to illustrate the workflow.
- Stage 2: The script runs on the actual model, producing real evaluation results along with cryptographic attestation.

The attestation process cryptographically binds the evaluation results to a TEE (Trusted Execution Environment) quote. This quote serves as proof that a specific model (identified by its hash) was executed within a legitimate TEE, and that the reported outputs are authentic and trustworthy. 

## Install lm-evaluation-harness
`lm-evaluation-harness` provides a unified framework to test generative language models on a large number of different evaluation tasks.

```python
!git clone --depth 1 https://github.com/EleutherAI/lm-evaluation-harness
%pip install -e ./lm-evaluation-harness[wandb]
```

## Model Selection（HuggingFace for Example）

```
HG_MODEL="deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B"
from lm_eval.utils import setup_logging
from lm_eval.models import huggingface
from lm_eval.tasks import TaskManager
lm_obj = huggingface.HFLM(pretrained=HG_MODEL, device="cpu")
task_manager = TaskManager()
lm_obj.get_model_info()
```

## Tasks Selection 

### BBQ: A Hand-Built Bias Benchmark for Question Answering

BBQ measures the bias in the output for the question answering task. The dataset of question-sets constructed by the authors that highlight attested social biases against people belonging to protected classes along nine social dimensions relevant for U.S. English-speaking contexts. BBQ evaluates model responses at two levels: (i) given an under-informative context, how strongly responses reflect social biases (AMBIGUOUS CONTEXT), and (ii) given an adequately informative context, whether the model's biases override a correct answer choice (DISAMBIGUATED CONTEXT).

### CrowS-Pairs: A Challenge Dataset for Measuring Social Biases in Masked Language Models

CrowS-Pairs is a challenge set for evaluating what language models (LMs) on their tendency to generate biased outputs. CrowS-Pairs comes in 2 languages and the English subset has a newer version which fixes some of the issues with the original version

### Simple Cooccurrence Bias

This bias evaluation relies on simple templates for prompting LMs and tests for bias in the next word prediction. For instance, when given a context such as "The {occupation} was a", masculine gender identifiers are found to be more likely to follow than feminine gender ones. Following Brown et al. (2020), this occupation bias is measured as the average log-likelihood of choosing a female gender identifier (woman, female) minus the log-likelihood of choosing a male gender identifier (man, male).

### Winogender: Gender Bias in Coreference Resolution
Winogender is designed to measure gender bias in coreference resolution systems, but has also been used for evaluating language models. The dataset consists of simple sentences with an occupation, participant, and pronoun, where the pronoun refers to either the occupation or participant. Each example consists of three variations, where only the gender of the pronoun is changed, to test how the pronoun affects the prediction. An example of the Winogender schema is "The paramedic performed CPR on the passenger even though he/she/they knew it was too late." This implementation follows the description from the paper "Language Models are Few-Shot Learners", which uses prompts.

```python
import datasets
datasets.config.HF_DATASETS_TRUST_REMOTE_CODE = True
tee_llm_evaluation_result = lm_eval.simple_evaluate( # call simple_evaluate
    model=lm_obj,
    tasks=["winogender","simple_cooccurrence_bias", "crows_pairs_english"],
    num_fewshot=0,
    task_manager=task_manager,
    log_samples=True,
    batch_size=1024,
    confirm_run_unsafe_code=True
)
tee_llm_evaluation_result["results"]
```

## Get Result and TEE Attestation Report
After the job finished, downloaded the result along with the attestation report. The `eat_nonce` in the attestation report is the hash of the output file.
