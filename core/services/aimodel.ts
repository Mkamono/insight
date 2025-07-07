import { google } from "@ai-sdk/google";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import { openai } from "@ai-sdk/openai";
import { anthropic } from "@ai-sdk/anthropic";
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

function getOpenaiModel(): LanguageModel {
    const apiKey = process.env.OPENAI_API_KEY;
    if (!apiKey) {
        throw new Error('OPENAI_API_KEY is required');
    }
    return openai(process.env.OPENAI_MODEL || 'gpt-4o-mini');
}

function getAnthropicModel(): LanguageModel {
    const apiKey = process.env.ANTHROPIC_API_KEY;
    if (!apiKey) {
        throw new Error('ANTHROPIC_API_KEY is required');
    }
    return anthropic(process.env.ANTHROPIC_MODEL || 'claude-3-haiku-20240307');
}

export function getAiModel(): LanguageModel {
    // 環境変数に基づいて自動選択（優先順位順）
    const providers = [
        {
            name: 'OpenAI Compatible',
            check: () => process.env.OPENAI_COMPATIBLE_API_KEY && process.env.OPENAI_COMPATIBLE_BASE_URL,
            create: () => getOpenaiCompatibleModel()
        },
        {
            name: 'OpenAI',
            check: () => process.env.OPENAI_API_KEY,
            create: () => getOpenaiModel()
        },
        {
            name: 'Anthropic',
            check: () => process.env.ANTHROPIC_API_KEY,
            create: () => getAnthropicModel()
        },
        {
            name: 'Google AI',
            check: () => process.env.GOOGLE_GENERATIVE_AI_API_KEY,
            create: () => getGoogleAiModel()
        }
    ];
    
    for (const provider of providers) {
        if (provider.check()) {
            console.log(`Using ${provider.name} API`);
            return provider.create();
        }
    }
    
    // どのプロバイダも設定されていない場合はエラー
    throw new Error(`No AI provider configured. Please set one of the following environment variables:
- OPENAI_COMPATIBLE_API_KEY + OPENAI_COMPATIBLE_BASE_URL
- OPENAI_API_KEY  
- ANTHROPIC_API_KEY
- GOOGLE_GENERATIVE_AI_API_KEY`);
}
