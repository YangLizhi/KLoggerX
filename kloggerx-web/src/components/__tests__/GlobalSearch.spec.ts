import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import GlobalSearch from '@/components/common/GlobalSearch.vue'

// Mock external dependencies
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/api/modules/document', () => ({
  searchDocuments: vi.fn().mockResolvedValue({ data: { items: [], total: 0 } }),
}))

describe('GlobalSearch', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('renders when visible is true', () => {
    const wrapper = mount(GlobalSearch, {
      props: {
        visible: true,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
        stubs: {
          'el-dialog': {
            template: '<div v-if="modelValue"><slot /></div>',
            props: ['modelValue'],
          },
          'el-icon': { template: '<span><slot /></span>' },
        },
        components: {
          Search: { template: '<i />' },
          Clock: { template: '<i />' },
          Close: { template: '<i />' },
          Loading: { template: '<i />' },
        },
      },
    })
    expect(wrapper.exists()).toBe(true)
  })

  it('has an input field for search', () => {
    const wrapper = mount(GlobalSearch, {
      props: {
        visible: true,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
        stubs: {
          'el-dialog': {
            template: '<div v-if="modelValue"><slot /></div>',
            props: ['modelValue'],
          },
          'el-icon': { template: '<span><slot /></span>' },
        },
        components: {
          Search: { template: '<i />' },
          Clock: { template: '<i />' },
          Close: { template: '<i />' },
          Loading: { template: '<i />' },
        },
      },
    })
    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
  })

  it('emits close on ESC keydown', async () => {
    const wrapper = mount(GlobalSearch, {
      props: {
        visible: true,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
        stubs: {
          'el-dialog': {
            template: '<div v-if="modelValue"><slot /></div>',
            props: ['modelValue'],
          },
          'el-icon': { template: '<span><slot /></span>' },
        },
        components: {
          Search: { template: '<i />' },
          Clock: { template: '<i />' },
          Close: { template: '<i />' },
          Loading: { template: '<i />' },
        },
      },
    })

    const input = wrapper.find('input')
    await input.trigger('keydown', { key: 'Escape' })
    await nextTick()

    expect(wrapper.emitted('update:visible')).toBeTruthy()
  })
})
