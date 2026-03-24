export function getTopicError(subject: string): string {
  if (!subject || subject === '') {
    return 'Topic is required';
  }

  const tokens = subject.split('.');
  if (tokens.some((token) => token === '>' || token === '*')) {
    return 'Wildcards are not allowed';
  }

  const isValidToken = (token: string): boolean => /^[^\s.]+$/.test(token);
  if (!tokens.every(isValidToken)) {
    return `Invalid topic: [${subject}]`;
  }

  return '';
}
