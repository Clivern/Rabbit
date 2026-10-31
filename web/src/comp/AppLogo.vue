<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :viewBox="variant === 'mask' ? '76 140 360 360' : '0 0 512 512'"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <template v-if="variant === 'hanging'">
      <g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
        <path d="M16 70Q256 44 496 74" stroke-width="20" />
        <path d="M156 210C134 160 132 118 144 82" stroke-width="40" />
        <path d="M356 210C378 160 380 118 368 84" stroke-width="40" />
        <path d="M128 92l-8-34M144 88v-36M160 92l8-34" stroke-width="10" />
        <path d="M352 94l-8-34M368 90v-36M384 94l8-34" stroke-width="10" />
      </g>
      <path d="M440 70q24 34 64 26q-14-42-64-26z" />
      <g transform="translate(77 134) scale(.7)">
        <path
          fill-rule="evenodd"
          d="M16 256A240 216 0 1 0 496 256A240 216 0 1 0 16 256ZM52 286A204 156 0 1 0 460 286A204 156 0 1 0 52 286ZM97.1 312.6A78 42 -28 1 0 234.9 239.4A78 42 -28 1 0 97.1 312.6ZM277.1 239.4A78 42 28 1 0 414.9 312.6A78 42 28 1 0 277.1 239.4ZM167 264A17 17 0 1 0 201 264A17 17 0 1 0 167 264ZM311 264A17 17 0 1 0 345 264A17 17 0 1 0 311 264ZM222 336A34 22 0 1 0 290 336A34 22 0 1 0 222 336Z"
        />
        <path
          d="M214 382Q256 410 298 382"
          fill="none"
          stroke="currentColor"
          stroke-width="16"
          stroke-linecap="round"
        />
      </g>
    </template>
    <template v-else-if="variant === 'mask'">
      <path
        fill-rule="evenodd"
        d="M97.1 312.6A78 42 -28 1 0 234.9 239.4A78 42 -28 1 0 97.1 312.6ZM277.1 239.4A78 42 28 1 0 414.9 312.6A78 42 28 1 0 277.1 239.4ZM167 264A17 17 0 1 0 201 264A17 17 0 1 0 167 264ZM311 264A17 17 0 1 0 345 264A17 17 0 1 0 311 264ZM222 336A34 22 0 1 0 290 336A34 22 0 1 0 222 336Z"
      />
      <path
        d="M208 384Q256 418 304 384"
        fill="none"
        stroke="currentColor"
        stroke-width="20"
        stroke-linecap="round"
      />
    </template>
    <template v-else-if="variant === 'tile'">
      <rect width="512" height="512" rx="116" />
      <g transform="translate(0 -30)">
        <ellipse cx="256" cy="286" rx="204" ry="156" fill="var(--theme-bg)" />
        <ellipse cx="166" cy="276" rx="78" ry="42" transform="rotate(-28 166 276)" />
        <ellipse cx="346" cy="276" rx="78" ry="42" transform="rotate(28 346 276)" />
        <circle cx="184" cy="264" r="17" fill="var(--theme-bg)" />
        <circle cx="328" cy="264" r="17" fill="var(--theme-bg)" />
        <ellipse cx="256" cy="336" rx="34" ry="22" />
        <path
          d="M214 382Q256 410 298 382"
          fill="none"
          stroke="currentColor"
          stroke-width="16"
          stroke-linecap="round"
        />
      </g>
    </template>
    <template v-else>
      <path
        fill-rule="evenodd"
        d="M16 256A240 216 0 1 0 496 256A240 216 0 1 0 16 256ZM52 286A204 156 0 1 0 460 286A204 156 0 1 0 52 286ZM97.1 312.6A78 42 -28 1 0 234.9 239.4A78 42 -28 1 0 97.1 312.6ZM277.1 239.4A78 42 28 1 0 414.9 312.6A78 42 28 1 0 277.1 239.4ZM167 264A17 17 0 1 0 201 264A17 17 0 1 0 167 264ZM311 264A17 17 0 1 0 345 264A17 17 0 1 0 311 264ZM222 336A34 22 0 1 0 290 336A34 22 0 1 0 222 336Z"
      />
      <path
        d="M214 382Q256 410 298 382"
        fill="none"
        stroke="currentColor"
        stroke-width="16"
        stroke-linecap="round"
      />
    </template>
  </svg>
</template>

<script setup>
defineProps({
  alt: {
    type: String,
    default: '',
  },
  variant: {
    type: String,
    default: 'current',
  },
})
</script>
