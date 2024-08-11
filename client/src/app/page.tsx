'use client';

import { useState } from 'react';

export default function Home() {
  const [url, setUrl] = useState('');
  const [shortUrl, setShortUrl] = useState('');

  const handleSubmit = async (e: any) => {
    e.preventDefault();

    try {
      const response = await fetch('http://localhost:4000/short-it', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: new URLSearchParams({ URL: url }),
      });

      if (!response.ok) {
        throw new Error('Failed to shorten URL');
      }

      const result = await response.text();
      setShortUrl(result);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <main className='flex min-h-screen flex-col items-center justify-between p-24'>
      URL shortener.
      <form onSubmit={handleSubmit}>
        <p>Shorten a URL:</p>
        <input
          type='text'
          placeholder='Enter URL'
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
        <button type='submit' className='bg-red-500 p-2 ml-3'>
          Shorten
        </button>
      </form>
      {shortUrl && (
        <p>
          Shortened URL:{' '}
          <a href={shortUrl} target='_blank' rel='noopener noreferrer'>
            {shortUrl}
          </a>
        </p>
      )}
    </main>
  );
}
