// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

export const handler = async (event: any) => {
  return {
    statusCode: 200,
    body: JSON.stringify({ message: 'serverless-ts running' }),
  };
};
