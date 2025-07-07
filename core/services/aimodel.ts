import { google } from "@ai-sdk/google";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import { LanguageModel } from "ai";

function getGoogleAiModel(): LanguageModel {
    const apiKey = process.env.GOOGLE_GENERATIVE_AI_API_KEY;
    if (!apiKey) {
        throw new Error('GOOGLE_GENERATIVE_AI_API_KEY');
    }
    return google('models/gemini-2.5-flash');
}

function getOpenaiCompatibleModel(): LanguageModel {
    const provider = createOpenAICompatible({
        name: 'OpenAI-Compatible',
        apiKey: process.env.OPENAI_COMPATIBLE_API_KEY,
        baseURL: process.env.OPENAI_COMPATIBLE_BASE_URL || '',
    });

    return provider(process.env.OPENAI_COMPATIBLE_MODEL || 'gpt-3.5-turbo');
}

export function getAiModel(): LanguageModel {
    // const model = getGoogleAiModel();
    const model = getOpenaiCompatibleModel();
    return model;
}
