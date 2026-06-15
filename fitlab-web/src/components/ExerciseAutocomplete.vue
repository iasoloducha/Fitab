<template>
  <div class="autocomplete-wrapper" ref="wrapper">
    <input
      type="text"
      :value="displayText"
      @input="onInput"
      @focus="onFocus"
      @blur="onBlur"
      @keydown.down.prevent="highlightNext"
      @keydown.up.prevent="highlightPrev"
      @keydown.enter.prevent="selectHighlighted"
      @keydown.escape="closeDropdown"
      :placeholder="placeholder"
      class="autocomplete-input"
    />
    <ul v-if="showDropdown && filteredExercises.length > 0" class="autocomplete-dropdown">
      <li
        v-for="(ex, index) in filteredExercises"
        :key="ex.id"
        :class="{ highlighted: index === highlightedIndex }"
        @mousedown.prevent="selectExercise(ex)"
        class="autocomplete-item"
      >
        {{ ex.name }}
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { api } from '../api'

const props = defineProps({
  modelValue: {
    type: Number,
    default: null,
  },
  placeholder: {
    type: String,
    default: 'Buscar ejercicio...',
  },
})

const emit = defineEmits(['update:modelValue', 'update:name'])

const wrapper = ref(null)
const inputValue = ref('')
const selectedName = ref('')
const selectedId = ref(props.modelValue)
const exercises = ref([])
const showDropdown = ref(false)
const highlightedIndex = ref(-1)
let debounceTimer = null

const displayText = computed(() => {
  if (selectedId.value && selectedName.value) {
    return selectedName.value
  }
  return inputValue.value
})

const filteredExercises = computed(() => {
  const q = inputValue.value.toLowerCase().trim()
  if (!q) return exercises.value
  return exercises.value.filter((ex) => ex.name.toLowerCase().includes(q))
})

watch(
  () => props.modelValue,
  (newVal) => {
    selectedId.value = newVal
    if (!newVal) {
      selectedName.value = ''
    }
  }
)

async function searchExercises(q) {
  try {
    const response = await api.catalog.list(q || '')
    exercises.value = response.data || []
  } catch {
    exercises.value = []
  }
}

function onInput(e) {
  const val = e.target.value
  inputValue.value = val
  selectedId.value = null
  selectedName.value = ''
  emit('update:modelValue', null)
  emit('update:name', val)

  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    searchExercises(val)
    showDropdown.value = true
    highlightedIndex.value = -1
  }, 300)
}

function onFocus() {
  if (filteredExercises.value.length > 0) {
    showDropdown.value = true
  } else {
    searchExercises(inputValue.value)
    showDropdown.value = true
  }
}

function onBlur() {
  setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

function selectExercise(ex) {
  selectedId.value = ex.id
  selectedName.value = ex.name
  inputValue.value = ex.name
  showDropdown.value = false
  emit('update:modelValue', ex.id)
  emit('update:name', ex.name)
}

function selectHighlighted() {
  if (highlightedIndex.value >= 0 && highlightedIndex.value < filteredExercises.value.length) {
    selectExercise(filteredExercises.value[highlightedIndex.value])
  }
}

function highlightNext() {
  if (filteredExercises.value.length === 0) return
  highlightedIndex.value = (highlightedIndex.value + 1) % filteredExercises.value.length
}

function highlightPrev() {
  if (filteredExercises.value.length === 0) return
  highlightedIndex.value =
    highlightedIndex.value <= 0
      ? filteredExercises.value.length - 1
      : highlightedIndex.value - 1
}

function closeDropdown() {
  showDropdown.value = false
}
</script>

<style scoped>
.autocomplete-wrapper {
  position: relative;
  width: 100%;
}

.autocomplete-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #2a2a2a;
  color: white;
  box-sizing: border-box;
}

.autocomplete-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.autocomplete-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  max-height: 200px;
  overflow-y: auto;
  background: #363636;
  border: 1px solid #444;
  border-top: none;
  border-radius: 0 0 4px 4px;
  z-index: 100;
  list-style: none;
  margin: 0;
  padding: 0;
}

.autocomplete-item {
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  color: #ccc;
}

.autocomplete-item:hover,
.autocomplete-item.highlighted {
  background: var(--color-primary);
  color: white;
}
</style>
