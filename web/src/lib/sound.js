// Short chat chimes made with the Web Audio API, so no sound files ship with the app.

let context = null

// Browsers only start audio after a user gesture, so the first call should come from one (the send click).
function audio() {
  context ??= new AudioContext()
  if (context.state === 'suspended') context.resume()
  return context
}

// Plays one soft note that fades in and out, gliding to `to` Hz when given.
function tone(at, frequency, duration, to = frequency) {
  const ctx = audio()
  const oscillator = ctx.createOscillator()
  const gain = ctx.createGain()

  oscillator.type = 'sine'
  oscillator.frequency.setValueAtTime(frequency, at)
  oscillator.frequency.exponentialRampToValueAtTime(to, at + duration)

  gain.gain.setValueAtTime(0.0001, at)
  gain.gain.exponentialRampToValueAtTime(0.08, at + 0.015)
  gain.gain.exponentialRampToValueAtTime(0.0001, at + duration)

  oscillator.connect(gain).connect(ctx.destination)
  oscillator.start(at)
  oscillator.stop(at + duration)
}

// A quick rising pop when the user's message goes out.
export function playSend() {
  tone(audio().currentTime, 520, 0.12, 880)
}

// Two soft rising notes when the assistant starts answering.
export function playReply() {
  const at = audio().currentTime
  tone(at, 660, 0.16)
  tone(at + 0.09, 990, 0.22)
}
