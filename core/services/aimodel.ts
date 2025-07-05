import { google } from "@ai-sdk/google";
import { LanguageModel } from "ai";

function getGoogleAiModel(): LanguageModel {
    const apiKey = process.env.GOOGLE_GENERATIVE_AI_API_KEY;
    if (!apiKey) {
        throw new Error('GOOGLE_GENERATIVE_AI_API_KEY');
    }
    return google('models/gemini-2.5-flash');
}

export function getAiModel(): LanguageModel {
    const model = getGoogleAiModel();
    return model;
}
