#!/usr/bin/env node
/**
 * Generate PWA icon files from the SVG favicon
 * This script creates PNG icons for PWA manifest
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { createCanvas } from 'canvas';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const publicDir = path.join(__dirname, '..', 'public');

// Create canvas-based PNG icons
const sizes = [
  { size: 192, maskable: false },
  { size: 192, maskable: true },
  { size: 512, maskable: false },
  { size: 512, maskable: true },
];

// Brand color
const brandColor = '#1D9E75';
const maskableBgColor = '#FFFFFF';

function generateIcon(size, isMaskable) {
  const canvas = createCanvas(size, size);
  const ctx = canvas.getContext('2d');

  // Background
  ctx.fillStyle = isMaskable ? maskableBgColor : brandColor;
  ctx.fillRect(0, 0, size, size);

  if (isMaskable) {
    // For maskable icons, add a smaller centered circle with the brand color
    const radius = size * 0.35;
    ctx.fillStyle = brandColor;
    ctx.beginPath();
    ctx.arc(size / 2, size / 2, radius, 0, Math.PI * 2);
    ctx.fill();

    // Add a lighter inner circle
    ctx.fillStyle = '#E1F5EE';
    const innerRadius = radius * 0.6;
    ctx.beginPath();
    ctx.arc(size / 2, size / 2, innerRadius, 0, Math.PI * 2);
    ctx.fill();
  } else {
    // For regular icons, add an "I" (Indexarr initial) in the center
    ctx.fillStyle = '#FFFFFF';
    ctx.font = `bold ${size * 0.5}px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText('I', size / 2, size / 2);
  }

  return canvas.toBuffer('image/png');
}

// Generate all icon sizes
try {
  for (const { size, maskable } of sizes) {
    const filename = maskable
      ? `pwa-${size}x${size}-maskable.png`
      : `pwa-${size}x${size}.png`;
    const filepath = path.join(publicDir, filename);

    const pngBuffer = generateIcon(size, maskable);
    fs.writeFileSync(filepath, pngBuffer);
    console.log(`✓ Generated ${filename} (${size}x${size})`);
  }

  // Also generate screenshot placeholders
  const screenshots = [
    { size: 540, height: 720, name: 'screenshot-540x720.png' },
    { size: 1280, height: 720, name: 'screenshot-1280x720.png' },
  ];

  for (const { size, height, name } of screenshots) {
    const canvas = createCanvas(size, height);
    const ctx = canvas.getContext('2d');

    // Gradient background
    const gradient = ctx.createLinearGradient(0, 0, 0, height);
    gradient.addColorStop(0, '#1D9E75');
    gradient.addColorStop(1, '#085041');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, size, height);

    // Add text
    ctx.fillStyle = '#FFFFFF';
    ctx.font = `bold ${Math.round(size * 0.1)}px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText('Indexarr', size / 2, height / 3);
    ctx.font = `${Math.round(size * 0.05)}px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`;
    ctx.fillText('Media Library Manager', size / 2, height / 2);

    const pngBuffer = canvas.toBuffer('image/png');
    fs.writeFileSync(path.join(publicDir, name), pngBuffer);
    console.log(`✓ Generated ${name} (${size}x${height})`);
  }

  console.log('\n✅ All PWA icons generated successfully!');
} catch (error) {
  console.error('❌ Error generating icons:', error);
  process.exit(1);
}
